package command

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/weaveworks/ignite/cmd/ignite/run"
	"github.com/weaveworks/ignite/pkg/apis/ignite"
	"github.com/weaveworks/ignite/pkg/apis/ignite/validation"
	meta "github.com/weaveworks/ignite/pkg/apis/meta/v1alpha1"
	"github.com/weaveworks/ignite/pkg/config"
	"github.com/weaveworks/ignite/pkg/constants"
	igniteNetwork "github.com/weaveworks/ignite/pkg/network"
	"github.com/weaveworks/ignite/pkg/operations"
	"github.com/weaveworks/ignite/pkg/preflight/checkers"
	"github.com/weaveworks/ignite/pkg/providers"
	"github.com/weaveworks/ignite/pkg/providers/client"
	"github.com/weaveworks/ignite/pkg/providers/manifeststorage"
	"github.com/weaveworks/ignite/pkg/providers/storage"
	igniteRuntime "github.com/weaveworks/ignite/pkg/runtime"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type commandRunner interface {
	RunCommand(ctx context.Context, command command, logger Logger) error
}

const firecrackerContainerDir = "/work"

var igniteProviderSync sync.Once

// formatFirecrackerCommand constructs the command to run on the host via a Firecracker
// virtual machine in order to invoke the given spec. If the spec specifies an image, then
// the command will be run inside of a container inside of the VM. Otherwise, the command
// will be run inside of the VM. The containers are one-shot and subject to the resource
// limits specified in the given options.
//
// The name value supplied here refers to the Firecracker virtual machine, which must have
// also been the name supplied to a successful invocation of setupFirecracker. Additionally,
// the virtual machine must not yet have been torn down (via teardownFirecracker).
func formatFirecrackerCommand(spec CommandSpec, name string, options Options) command {
	rawOrDockerCommand := formatRawOrDockerCommand(spec, firecrackerContainerDir, options)

	innerCommand := strings.Join(rawOrDockerCommand.Command, " ")
	if len(rawOrDockerCommand.Env) > 0 {
		// If we have env vars that are arguments to the command we need to escape them
		quotedEnv := quoteEnv(rawOrDockerCommand.Env)
		innerCommand = fmt.Sprintf("%s %s", strings.Join(quotedEnv, " "), innerCommand)
	}
	if rawOrDockerCommand.Dir != "" {
		innerCommand = fmt.Sprintf("cd %s && %s", rawOrDockerCommand.Dir, innerCommand)
	}

	return command{
		Key:       spec.Key,
		Command:   []string{"ignite", "exec", name, "--", innerCommand},
		Operation: spec.Operation,
	}
}

// setupFirecracker invokes a set of commands to provision and prepare a Firecracker virtual
// machine instance. If a startup script path (an executable file on the host) is supplied,
// it will be mounted into the new virtual machine instance and executed.
func setupFirecracker(ctx context.Context, runner commandRunner, logger Logger, name, repoDir string, options Options, op *Operations) error {
	// Start the VM and wait for the SSH server to become available
	_ = command{
		Key: "setup.firecracker.start",
		Command: flatten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "cni",
			// firecrackerResourceFlags(options.ResourceOptions),
			firecrackerCopyfileFlags(repoDir, options.FirecrackerOptions.VMStartupScriptPath),
			"--ssh",
			"--name", name,
			sanitizeImage(options.FirecrackerOptions.Image),
		),
		Operation: op.SetupFirecrackerStart,
	}

	igniteProviderSync.Do(func() {
		if err := manifeststorage.SetManifestStorage(); err != nil {
			panic(fmt.Sprintf("failed to set ignite manifest storage: %v", err))
		}
		if err := storage.SetGenericStorage(); err != nil {
			panic(fmt.Sprintf("failed to set generic ignite storage: %v", err))
		}
		if err := client.SetClient(); err != nil {
			panic(fmt.Sprintf("failed to set ignite client: %v", err))
		}
		if err := config.SetAndPopulateProviders(igniteRuntime.RuntimeDocker, igniteNetwork.PluginCNI); err != nil {
			panic(fmt.Sprintf("failed to populate ignite providers: %v", err))
		}
	})

	if err := callWithInstrumentedLock(op, logger, func() error {
		baseVM := providers.Client.VMs().New()

		baseVM.Name = name

		baseVM.Status.Runtime.Name = igniteRuntime.RuntimeDocker
		baseVM.Status.Network.Plugin = igniteNetwork.PluginCNI

		ociRef, err := meta.NewOCIImageRef(sanitizeImage(options.FirecrackerOptions.Image))
		if err != nil {
			return errors.Wrap(err, "failed to create OCI image ref")
		}
		baseVM.Spec.Image.OCI = ociRef
		baseVM.Spec.CPUs = uint64(options.ResourceOptions.NumCPUs)
		baseVM.Spec.Memory, _ = meta.NewSizeFromString(options.ResourceOptions.Memory)
		baseVM.Spec.DiskSize, _ = meta.NewSizeFromString(options.ResourceOptions.DiskSpace)

		if repoDir != "" {
			baseVM.Spec.CopyFiles = append(baseVM.Spec.CopyFiles, ignite.FileMapping{
				HostPath: repoDir,
				VMPath:   firecrackerContainerDir,
			})
		}
		if options.FirecrackerOptions.VMStartupScriptPath != "" {
			baseVM.Spec.CopyFiles = append(baseVM.Spec.CopyFiles, ignite.FileMapping{
				HostPath: options.FirecrackerOptions.VMStartupScriptPath,
				VMPath:   options.FirecrackerOptions.VMStartupScriptPath,
			})
		}

		baseVM.Spec.SSH = &ignite.SSH{Generate: true}

		if err := validation.ValidateVM(baseVM).ToAggregate(); err != nil {
			return errors.Wrap(err, "build invalid ignite vm")
		}

		createOpts := &run.CreateOptions{CreateFlags: &run.CreateFlags{
			CopyFiles:   firecrackerCopyfileFlags(repoDir, options.FirecrackerOptions.VMStartupScriptPath),
			SSH:         ignite.SSH{Generate: true},
			VM:          baseVM,
			RequireName: true,
		}}

		img, err := operations.FindOrImportImage(providers.Client, baseVM.Spec.Image.OCI)
		if err != nil {
			return errors.Wrap(err, "failed to import OCI image")
		}
		baseVM.SetImage(img)

		kernel, err := operations.FindOrImportKernel(providers.Client, baseVM.Spec.Kernel.OCI)
		if err != nil {
			return errors.Wrap(err, "failed to import kernel")
		}
		baseVM.SetKernel(kernel)

		if err := run.Create(createOpts); err != nil {
			return errors.Wrap(err, "failed to create vm")
		}

		if err := checkers.StartCmdChecks(baseVM, sets.String{}); err != nil {
			return errors.Wrap(err, "failed pre-start checks")
		}

		if err := operations.StartVM(baseVM, false); err != nil {
			return errors.Wrap(err, "failed to start ignite vm")
		}

		return waitForSSH(baseVM, constants.SSH_DEFAULT_TIMEOUT_SECONDS, constants.IGNITE_SPAWN_TIMEOUT)
	}); err != nil {
		return errors.Wrap(err, "failed to start firecracker vm")
	}

	if options.FirecrackerOptions.VMStartupScriptPath != "" {
		_ = command{
			Key:       "setup.startup-script",
			Command:   flatten("ignite", "exec", name, "--", options.FirecrackerOptions.VMStartupScriptPath),
			Operation: op.SetupStartupScript,
		}
		execOpts, err := (&run.ExecFlags{}).NewExecOptions(name, options.FirecrackerOptions.VMStartupScriptPath)
		if err != nil {
			return errors.Wrap(err, "failed to create exec options")
		}
		if err := run.Exec(execOpts); err != nil {
			return errors.Wrap(err, "failed to run startup script")
		}
	}

	return nil
}

// We've recently seen issues with concurent VM creation. It's likely we
// can do better here and run an empty VM at application startup, but I
// want to do this quick and dirty to see if we can raise our concurrency
// without other issues.
//
// https://github.com/weaveworks/ignite/issues/559
// Following up in https://github.com/sourcegraph/sourcegraph/issues/21377.
var igniteRunLock sync.Mutex

// callWithInstrumentedLock calls f while holding the igniteRunLock. The duration of the wait
// and active portions of this method are emitted as prometheus metrics.
func callWithInstrumentedLock(operations *Operations, logger Logger, f func() error) error {
	handle := logger.Log("setup.firecracker.runlock", nil)

	lockRequestedAt := time.Now()

	igniteRunLock.Lock()

	lockAcquiredAt := time.Now()

	handle.Finalize(0)
	handle.Close()

	err := f()

	lockReleasedAt := time.Now()

	igniteRunLock.Unlock()

	operations.RunLockWaitTotal.Add(float64(lockAcquiredAt.Sub(lockRequestedAt) / time.Millisecond))
	operations.RunLockHeldTotal.Add(float64(lockReleasedAt.Sub(lockAcquiredAt) / time.Millisecond))
	return err
}

func dialSuccess(vm *ignite.VM, seconds int) error {
	addr := vm.Status.Network.IPAddresses[0].String() + ":22"
	const perSecond = 10
	delay := time.Second / time.Duration(perSecond)
	var err error
	for i := 0; i < seconds*perSecond; i++ {
		conn, dialErr := net.DialTimeout("tcp", addr, delay)
		if conn != nil {
			conn.Close()
			err = nil
			break
		}
		err = dialErr
		time.Sleep(delay)
	}
	if err != nil {
		if err, ok := err.(*net.OpError); ok && err.Timeout() {
			return errors.Newf("tried connecting to SSH but timed out %s", err)
		}
		return err
	}

	return nil
}

func waitForSSH(vm *ignite.VM, dialSeconds int, sshTimeout time.Duration) error {
	if err := dialSuccess(vm, dialSeconds); err != nil {
		return err
	}

	certCheck := &ssh.CertChecker{
		IsHostAuthority: func(auth ssh.PublicKey, address string) bool {
			return true
		},
		IsRevoked: func(cert *ssh.Certificate) bool {
			return false
		},
		HostKeyFallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	config := &ssh.ClientConfig{
		HostKeyCallback: certCheck.CheckHostKey,
		Timeout:         sshTimeout,
	}

	addr := vm.Status.Network.IPAddresses[0].String() + ":22"
	sshConn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		if strings.Contains(err.Error(), "unable to authenticate") {
			// we connected to the ssh server and recieved the expected failure
			return nil
		}
		return err
	}

	defer sshConn.Close()
	return errors.New("connected successfully with no authentication, failure was expected")
}

// teardownFirecracker issues a stop and a remove request for the Firecracker VM with
// the given name.
func teardownFirecracker(ctx context.Context, runner commandRunner, logger Logger, name string, operations *Operations) error {
	removeCommand := command{
		Key:       "teardown.firecracker.remove",
		Command:   flatten("ignite", "rm", "-f", name),
		Operation: operations.TeardownFirecrackerRemove,
	}
	if err := runner.RunCommand(ctx, removeCommand, logger); err != nil {
		log15.Error("Failed to remove firecracker vm", "name", name, "err", err)
	}

	return nil
}

func firecrackerCopyfileFlags(dir, vmStartupScriptPath string) []string {
	copyfiles := make([]string, 0, 2)
	if dir != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", dir, firecrackerContainerDir))
	}
	if vmStartupScriptPath != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", vmStartupScriptPath, vmStartupScriptPath))
	}

	sort.Strings(copyfiles)
	return intersperse("--copy-files", copyfiles)
}

var imagePattern = lazyregexp.New(`([^:@]+)(?::([^@]+))?(?:@sha256:([a-z0-9]{64}))?`)

// sanitizeImage sanitizes the given docker image for use by ignite. The ignite utility
// has some issue parsing docker tags that include a sha256 hash, so we try to remove it
// from any of the image references before passing it to the ignite command.
func sanitizeImage(image string) string {
	if matches := imagePattern.FindStringSubmatch(image); len(matches) == 4 {
		if matches[2] == "" {
			return matches[1]
		}

		return fmt.Sprintf("%s:%s", matches[1], matches[2])
	}

	return image
}
