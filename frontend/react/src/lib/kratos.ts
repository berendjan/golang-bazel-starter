import { Configuration, FrontendApi } from '@ory/client';

// Kratos public API client
// Uses same-origin proxy to avoid CORS/CSRF issues with cross-origin cookies
// Set VITE_KRATOS_URL environment variable to override (e.g., for local dev)
const kratosBasePath = import.meta.env.VITE_KRATOS_URL || '/kratos';

const kratosConfig = new Configuration({
  basePath: kratosBasePath,
  baseOptions: {
    withCredentials: true,
  },
});

export const kratos = new FrontendApi(kratosConfig);

// Helper to extract error messages from Kratos UI nodes
export function getErrorMessages(ui: { nodes?: Array<{ messages?: Array<{ text: string }> }> }): string[] {
  const messages: string[] = [];
  if (ui.nodes) {
    for (const node of ui.nodes) {
      if (node.messages) {
        for (const message of node.messages) {
          messages.push(message.text);
        }
      }
    }
  }
  return messages;
}
