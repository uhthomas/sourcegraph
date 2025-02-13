import { CodebaseContext } from '../../codebase-context'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_RECIPE_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { getShortTimestamp } from '../../timestamp'
import { Interaction } from '../transcript/interaction'

import { languageMarkdownID, languageNames } from './langs'
import { Recipe } from './recipe'

export class TranslateToLanguage implements Recipe {
    public getID(): string {
        return 'translate-to-language'
    }

    public async getInteraction(
        _humanChatInput: string,
        editor: Editor,
        _intentDetector: IntentDetector,
        _codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const selection = editor.getActiveTextEditorSelection()
        if (!selection) {
            return null
        }

        const toLanguage = await editor.showQuickPick(languageNames)
        if (!toLanguage) {
            // TODO: Show the warning within the Chat UI.
            // editor.showWarningMessage('Must pick a language to translate to.')
            return null
        }

        const timestamp = getShortTimestamp()
        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_INPUT_TOKENS)

        const promptMessage = `Translate the following code into ${toLanguage}\n\`\`\`\n${truncatedSelectedText}\n\`\`\``
        const displayText = `Translate the following code into ${toLanguage}\n\`\`\`\n${selection.selectedText}\n\`\`\``

        const markdownID = languageMarkdownID[toLanguage] || ''
        const assistantResponsePrefix = `Here is the code translated to ${toLanguage}:\n\`\`\`${markdownID}\n`

        return new Interaction(
            { speaker: 'human', text: promptMessage, displayText, timestamp },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
                displayText: '',
                timestamp,
            },
            Promise.resolve([])
        )
    }
}
