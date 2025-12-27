import { useEffect, useState, FormEvent } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { LoginFlow, UiNode, UiNodeInputAttributes, UpdateLoginFlowBody } from '@ory/client';
import { kratos, getErrorMessages } from '../lib/kratos';
import { useAuth } from '../context/AuthContext';

export function Login() {
  const [flow, setFlow] = useState<LoginFlow | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { session, refreshSession } = useAuth();

  // If already logged in, redirect to dashboard
  useEffect(() => {
    if (session) {
      navigate('/dashboard');
    }
  }, [session, navigate]);

  useEffect(() => {
    const flowId = searchParams.get('flow');

    const initFlow = async () => {
      try {
        setLoading(true);
        if (flowId) {
          // Resume existing flow
          const { data } = await kratos.getLoginFlow({ id: flowId });
          setFlow(data);
        } else {
          // Create new flow
          const { data } = await kratos.createBrowserLoginFlow();
          // Update URL with flow ID for resumability
          window.history.replaceState(null, '', `?flow=${data.id}`);
          setFlow(data);
        }
      } catch (err) {
        console.error('Failed to initialize login flow:', err);
        setError('Failed to initialize login. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    initFlow();
  }, [searchParams]);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!flow) return;

    setSubmitting(true);
    setError(null);

    const formData = new FormData(e.currentTarget);
    const body: Record<string, string> = {};
    formData.forEach((value, key) => {
      body[key] = value.toString();
    });

    try {
      await kratos.updateLoginFlow({
        flow: flow.id,
        updateLoginFlowBody: {
          method: 'password',
          ...body,
        } as UpdateLoginFlowBody,
      });
      // Login successful, refresh session and redirect
      await refreshSession();
      navigate('/dashboard');
    } catch (err) {
      const response = (err as { response?: { data?: LoginFlow } })?.response?.data;
      if (response) {
        setFlow(response);
        const messages = getErrorMessages(response.ui);
        if (messages.length > 0) {
          setError(messages.join(', '));
        }
      } else {
        setError('Login failed. Please try again.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div style={styles.container}>
        <p>Loading...</p>
      </div>
    );
  }

  const inputNodes = flow?.ui.nodes.filter(
    (node) => node.group === 'default' || node.group === 'password'
  ) ?? [];

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h1 style={styles.title}>Login</h1>

        {error && <div style={styles.error}>{error}</div>}

        <form onSubmit={handleSubmit} style={styles.form}>
          {inputNodes.map((node: UiNode) => {
            const attrs = node.attributes as UiNodeInputAttributes;
            if (attrs.type === 'hidden') {
              return <input key={attrs.name} type="hidden" name={attrs.name} value={attrs.value as string} />;
            }
            if (attrs.type === 'submit') {
              return (
                <button
                  key={attrs.name}
                  type="submit"
                  name={attrs.name}
                  value={attrs.value as string}
                  disabled={submitting}
                  style={styles.button}
                >
                  {submitting ? 'Signing in...' : 'Sign In'}
                </button>
              );
            }
            return (
              <div key={attrs.name} style={styles.field}>
                <label style={styles.label}>{node.meta.label?.text || attrs.name}</label>
                <input
                  type={attrs.type}
                  name={attrs.name}
                  defaultValue={attrs.value as string}
                  required={attrs.required}
                  disabled={attrs.disabled}
                  style={styles.input}
                />
              </div>
            );
          })}
        </form>

        <div style={styles.links}>
          <Link to="/auth/register" style={styles.link}>
            Don't have an account? Register
          </Link>
          <Link to="/auth/recovery" style={styles.link}>
            Forgot password?
          </Link>
        </div>
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    backgroundColor: '#f5f5f5',
  },
  card: {
    backgroundColor: 'white',
    padding: '2rem',
    borderRadius: '8px',
    boxShadow: '0 2px 10px rgba(0, 0, 0, 0.1)',
    width: '100%',
    maxWidth: '400px',
  },
  title: {
    textAlign: 'center',
    marginBottom: '1.5rem',
    color: '#333',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '1rem',
  },
  field: {
    display: 'flex',
    flexDirection: 'column',
    gap: '0.5rem',
  },
  label: {
    fontWeight: 500,
    color: '#555',
  },
  input: {
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '1rem',
  },
  button: {
    padding: '0.75rem',
    backgroundColor: '#007bff',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    fontSize: '1rem',
    cursor: 'pointer',
    marginTop: '1rem',
  },
  error: {
    backgroundColor: '#fee',
    color: '#c00',
    padding: '0.75rem',
    borderRadius: '4px',
    marginBottom: '1rem',
    textAlign: 'center',
  },
  links: {
    marginTop: '1.5rem',
    display: 'flex',
    flexDirection: 'column',
    gap: '0.5rem',
    textAlign: 'center',
  },
  link: {
    color: '#007bff',
    textDecoration: 'none',
  },
};
