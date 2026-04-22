export interface SuggestionModel {
    id: string;
    name: string;
    provider: string;
}

export const getAllModels = (): SuggestionModel[] => {
    return [
        { id: "gemini-2.0-flash-exp", name: "Gemini 2.0 Flash Exp", provider: "Google" },
        { id: "gemini-1.5-pro", name: "Gemini 1.5 Pro", provider: "Google" },
        { id: "gemini-1.5-flash", name: "Gemini 1.5 Flash", provider: "Google" },

        { id: "gpt-4o", name: "GPT-4o", provider: "OpenAI" },
        { id: "gpt-4o-mini", name: "GPT-4o Mini", provider: "OpenAI" },
        { id: "o1-preview", name: "o1 Preview", provider: "OpenAI" },
        { id: "o1-mini", name: "o1 Mini", provider: "OpenAI" },

        { id: "claude-3-5-sonnet-latest", name: "Claude 3.5 Sonnet", provider: "Anthropic" },
        { id: "claude-3-5-haiku-latest", name: "Claude 3.5 Haiku", provider: "Anthropic" },
        { id: "claude-3-opus-latest", name: "Claude 3 Opus", provider: "Anthropic" },

        { id: "deepseek-chat", name: "DeepSeek V3", provider: "DeepSeek" },
        { id: "deepseek-reasoner", name: "DeepSeek R1", provider: "DeepSeek" }
    ];
};
