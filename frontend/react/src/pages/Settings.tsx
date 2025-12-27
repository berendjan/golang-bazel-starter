import { useEffect, useState, FormEvent } from 'react';
import { useSearchParams, Link, useNavigate } from 'react-router-dom';
import { SettingsFlow, UiNode, UiNodeInputAttributes, UpdateSettingsFlowBody } from '@ory/client';
import { kratos, getErrorMessages } from '../lib/kratos';
import { useAuth } from '../context/AuthContext';

export function Settings() {
  const [flow, setFlow] = useState<SettingsFlow | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [searchParams] = useSearchParams();
  const { session, refreshSession } = useAuth();
  const navigate = useNavigate();

  // Redirect if not logged in
  useEffect(() => {
    if (!session && !loading) {
      navigate('/auth/login');
    }
  }, [session, loading, navigate]);

  useEffect(() => {
    const flowId = searchParams.get('flow');

    const initFlow = async () => {
      try {
        setLoading(true);
        if (flowId) {
          const { data } = await kratos.getSettingsFlow({ id: flowId });
          setFlow(data);
        } else {
          const { data } = await kratos.createBrowserSettingsFlow();
          window.history.replaceState(null, '', `?flow=${data.id}`);
          setFlow(data);
        }
      } catch (err) {
        console.error('Failed to initialize settings flow:', err);
        setError('Failed to load settings. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    initFlow();
  }, [searchParams]);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>, method: string) => {
    e.preventDefault();
    if (!flow) return;

    setSubmitting(true);
    setError(null);
    setSuccess(null);

    const formData = new FormData(e.currentTarget);
    const body: Record<string, string> = {};
    formData.forEach((value, key) => {
      body[key] = value.toString();
    });

    try {
      const { data } = await kratos.updateSettingsFlow({
        flow: flow.id,
        updateSettingsFlowBody: {
          method,
          ...body,
        } as UpdateSettingsFlowBody,
      });
      setFlow(data);
      setSuccess('Settings updated successfully!');
      await refreshSession();
    } catch (err) {
      const response = (err as { response?: { data?: SettingsFlow } })?.response?.data;
      if (response) {
        setFlow(response);
        const messages = getErrorMessages(response.ui);
        if (messages.length > 0) {
          setError(messages.join(', '));
        }
      } else {
        setError('Failed to update settings. Please try again.');
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

  // Group nodes by their group
  const profileNodes = flow?.ui.nodes.filter((node) => node.group === 'profile') ?? [];
  const passwordNodes = flow?.ui.nodes.filter((node) => node.group === 'password') ?? [];
  const defaultNodes = flow?.ui.nodes.filter((node) => node.group === 'default') ?? [];

  const renderNodes = (nodes: UiNode[], buttonText: string) => {
    return nodes.map((node: UiNode) => {
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
            {submitting ? 'Saving...' : buttonText}
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
    });
  };

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h1 style={styles.title}>Account Settings</h1>

        {error && <div style={styles.error}>{error}</div>}
        {success && <div style={styles.success}>{success}</div>}

        {profileNodes.length > 0 && (
          <div style={styles.section}>
            <h2 style={styles.sectionTitle}>Profile</h2>
            <form onSubmit={(e) => handleSubmit(e, 'profile')} style={styles.form}>
              {renderNodes([...defaultNodes, ...profileNodes], 'Update Profile')}
            </form>
          </div>
        )}

        {passwordNodes.length > 0 && (
          <div style={styles.section}>
            <h2 style={styles.sectionTitle}>Change Password</h2>
            <form onSubmit={(e) => handleSubmit(e, 'password')} style={styles.form}>
              {renderNodes(passwordNodes, 'Update Password')}
            </form>
          </div>
        )}

        <div style={styles.links}>
          <Link to="/dashboard" style={styles.link}>
            Back to Dashboard
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
    padding: '1rem',
  },
  card: {
    backgroundColor: 'white',
    padding: '2rem',
    borderRadius: '8px',
    boxShadow: '0 2px 10px rgba(0, 0, 0, 0.1)',
    width: '100%',
    maxWidth: '500px',
  },
  title: {
    textAlign: 'center',
    marginBottom: '1.5rem',
    color: '#333',
  },
  section: {
    marginBottom: '2rem',
    paddingBottom: '2rem',
    borderBottom: '1px solid #eee',
  },
  sectionTitle: {
    fontSize: '1.2rem',
    marginBottom: '1rem',
    color: '#555',
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
    marginTop: '0.5rem',
  },
  error: {
    backgroundColor: '#fee',
    color: '#c00',
    padding: '0.75rem',
    borderRadius: '4px',
    marginBottom: '1rem',
    textAlign: 'center',
  },
  success: {
    backgroundColor: '#efe',
    color: '#060',
    padding: '0.75rem',
    borderRadius: '4px',
    marginBottom: '1rem',
    textAlign: 'center',
  },
  links: {
    marginTop: '1rem',
    textAlign: 'center',
  },
  link: {
    color: '#007bff',
    textDecoration: 'none',
  },
};
