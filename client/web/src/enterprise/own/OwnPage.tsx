import { mdiAccount, mdiOpenInNew } from '@mdi/js'

import {
    Container,
    H1,
    H2,
    H3,
    Icon,
    Link,
    PageHeader,
    ProductStatusBadge,
    Text,
    useReducedMotion,
} from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'

/**
 * A page explaining how to use Sourcegraph Own.
 * Note that this page is visible regardless of whether the `search-ownership`
 * feature flag has been enabled.
 */
export const OwnPage: React.FunctionComponent<{}> = () => {
    const allowAutoplay = !useReducedMotion()

    return (
        <Page>
            <PageTitle title="Sourcegraph Own" />
            <PageHeader description="Track and update code ownership across your entire codebase." className="mb-3">
                <H1 as="h2" className="d-flex align-items-center">
                    <Icon svgPath={mdiAccount} aria-hidden={true} />
                    <span className="ml-2">Own</span>
                    <ProductStatusBadge status="experimental" className="ml-2" />
                </H1>
            </PageHeader>

            <Container className="mb-3">
                <div className="row align-items-start">
                    <div className="col-12 col-md-7">
                        <video
                            className="w-100 h-auto shadow percy-hide"
                            width={1280}
                            height={720}
                            autoPlay={allowAutoplay}
                            muted={true}
                            loop={true}
                            playsInline={true}
                            controls={!allowAutoplay}
                        >
                            <source
                                type="video/mp4"
                                src="https://storage.googleapis.com/sourcegraph-assets/own-search-file-owners.mp4"
                            />
                        </video>
                    </div>
                    <div className="col-12 col-md-5">
                        <H2 as="h3">Evergreen code ownership data for your entire codebase</H2>
                        <Text>
                            CODEOWNERS alone is not enough. Own makes it easy to find the owner of any file within your
                            codebase.
                        </Text>
                        <H3 as="h4">Use Own to…</H3>
                        <ul>
                            <li>Search for vulnerable code and reach out to the owner in seconds</li>
                            <li>Find who to ask about unfamiliar code</li>
                            <li>Determine who to request code reviews from without having to ask around</li>
                        </ul>
                        <Text>Ingest ownership data from CODEOWNERS files, or from an existing ownership system.</Text>
                        <Text>
                            Own is currently an experimental feature and is getting smarter fast. We’d love your
                            feedback. You can turn it on using the documentation below, or{' '}
                            <Link to="https://about.sourcegraph.com/demo">contact us</Link> to get a demo and learn more
                            about our roadmap.
                        </Text>
                        <H3 as="h4">Resources</H3>
                        <ul>
                            <li>
                                <Link to="/help/own" target="_blank" rel="noopener">
                                    Documentation{' '}
                                    <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                                </Link>
                            </li>
                            <li>
                                <Link to="https://about.sourcegraph.com/own" target="_blank" rel="noopener">
                                    Product page{' '}
                                    <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                                </Link>
                            </li>
                        </ul>
                    </div>
                </div>
            </Container>
        </Page>
    )
}
