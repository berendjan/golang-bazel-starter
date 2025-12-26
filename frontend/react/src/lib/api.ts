// API client for the gRPC-gateway server
// Uses same-origin proxy to the grpcserver

const apiBasePath = '/api';

interface Account {
  account_id: {
    id: string;
  };
}

interface ListAccountsResponse {
  accounts: Account[];
}

interface AccountCreationRequest {
  name: string;
}

interface StatusResponse {
  success: boolean;
  message?: string;
}

class ApiClient {
  private basePath: string;

  constructor(basePath: string) {
    this.basePath = basePath;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const response = await fetch(`${this.basePath}${path}`, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || `Request failed with status ${response.status}`);
    }

    return response.json();
  }

  async listAccounts(): Promise<Account[]> {
    const response = await this.request<ListAccountsResponse>('GET', '/v1/accounts');
    return response.accounts || [];
  }

  async createAccount(name: string): Promise<Account> {
    const body: AccountCreationRequest = { name };
    return this.request<Account>('POST', '/v1/accounts', body);
  }

  async deleteAccount(id: string): Promise<StatusResponse> {
    return this.request<StatusResponse>('DELETE', `/v1/accounts/${id}`);
  }
}

export const api = new ApiClient(apiBasePath);
export type { Account };
