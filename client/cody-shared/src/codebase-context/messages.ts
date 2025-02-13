import { Message } from '../sourcegraph-api'

export interface ContextMessage extends Message {
    fileName?: string
}

export function getContextMessageWithResponse(
    text: string,
    fileName: string,
    response: string = 'Ok.'
): ContextMessage[] {
    return [
        { speaker: 'human', text, fileName },
        { speaker: 'assistant', text: response },
    ]
}
