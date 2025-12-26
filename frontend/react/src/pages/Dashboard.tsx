import { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';
import { api, Account } from '../lib/api';

export function Dashboard() {
  const { identity, logout, loading: authLoading } = useAuth();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newAccountName, setNewAccountName] = useState('');
  const [creating, setCreating] = useState(false);

  const fetchAccounts = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.listAccounts();
      setAccounts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load accounts');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAccounts();
  }, []);

  const handleCreateAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newAccountName.trim()) return;

    try {
      setCreating(true);
      setError(null);
      await api.createAccount(newAccountName.trim());
      setNewAccountName('');
      await fetchAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create account');
    } finally {
      setCreating(false);
    }
  };

  const handleDeleteAccount = async (id: string) => {
    if (!confirm('Are you sure you want to delete this account?')) return;

    try {
      setError(null);
      await api.deleteAccount(id);
      await fetchAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete account');
    }
  };

  if (authLoading) {
    return (
      <div style={styles.container}>
        <p>Loading...</p>
      </div>
    );
  }

  const email = identity?.traits?.email as string | undefined;

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <div style={styles.headerContent}>
          <h1 style={styles.title}>Dashboard</h1>
          <div style={styles.userInfo}>
            {email && <span style={styles.email}>{email}</span>}
            <Link to="/auth/settings" style={styles.settingsLink}>
              Settings
            </Link>
            <button onClick={logout} style={styles.logoutButton}>
              Sign Out
            </button>
          </div>
        </div>
      </div>

      <div style={styles.content}>
        {error && <div style={styles.error}>{error}</div>}

        <div style={styles.card}>
          <h2 style={styles.cardTitle}>Create Account</h2>
          <form onSubmit={handleCreateAccount} style={styles.form}>
            <input
              type="text"
              value={newAccountName}
              onChange={(e) => setNewAccountName(e.target.value)}
              placeholder="Account name"
              style={styles.input}
              disabled={creating}
            />
            <button
              type="submit"
              style={styles.createButton}
              disabled={creating || !newAccountName.trim()}
            >
              {creating ? 'Creating...' : 'Create'}
            </button>
          </form>
        </div>

        <div style={styles.card}>
          <h2 style={styles.cardTitle}>Accounts</h2>
          {loading ? (
            <p>Loading accounts...</p>
          ) : accounts.length === 0 ? (
            <p style={styles.emptyState}>No accounts yet. Create one above!</p>
          ) : (
            <table style={styles.table}>
              <thead>
                <tr>
                  <th style={styles.th}>Account ID</th>
                  <th style={styles.th}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {accounts.map((account) => (
                  <tr key={account.account_id.id} style={styles.tr}>
                    <td style={styles.td}>
                      <code>{account.account_id.id}</code>
                    </td>
                    <td style={styles.td}>
                      <button
                        onClick={() => handleDeleteAccount(account.account_id.id)}
                        style={styles.deleteButton}
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    minHeight: '100vh',
    backgroundColor: '#f5f5f5',
  },
  header: {
    backgroundColor: 'white',
    borderBottom: '1px solid #ddd',
    padding: '1rem 2rem',
  },
  headerContent: {
    maxWidth: '1200px',
    margin: '0 auto',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  title: {
    margin: 0,
    color: '#333',
  },
  userInfo: {
    display: 'flex',
    alignItems: 'center',
    gap: '1rem',
  },
  email: {
    color: '#666',
  },
  settingsLink: {
    color: '#007bff',
    textDecoration: 'none',
  },
  logoutButton: {
    padding: '0.5rem 1rem',
    backgroundColor: '#dc3545',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
  },
  content: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '2rem',
  },
  error: {
    backgroundColor: '#fee',
    color: '#c00',
    padding: '1rem',
    borderRadius: '4px',
    marginBottom: '1rem',
  },
  card: {
    backgroundColor: 'white',
    borderRadius: '8px',
    boxShadow: '0 2px 10px rgba(0, 0, 0, 0.1)',
    padding: '1.5rem',
    marginBottom: '1.5rem',
  },
  cardTitle: {
    margin: '0 0 1rem 0',
    color: '#333',
  },
  form: {
    display: 'flex',
    gap: '1rem',
  },
  input: {
    flex: 1,
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '1rem',
  },
  createButton: {
    padding: '0.75rem 1.5rem',
    backgroundColor: '#28a745',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    fontSize: '1rem',
    cursor: 'pointer',
  },
  table: {
    width: '100%',
    borderCollapse: 'collapse',
  },
  th: {
    textAlign: 'left',
    padding: '0.75rem',
    borderBottom: '2px solid #ddd',
    color: '#666',
  },
  tr: {
    borderBottom: '1px solid #eee',
  },
  td: {
    padding: '0.75rem',
  },
  deleteButton: {
    padding: '0.25rem 0.75rem',
    backgroundColor: '#dc3545',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    fontSize: '0.875rem',
    cursor: 'pointer',
  },
  emptyState: {
    color: '#666',
    fontStyle: 'italic',
  },
};
