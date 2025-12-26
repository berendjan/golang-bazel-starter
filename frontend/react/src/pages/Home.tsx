import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export function Home() {
  const { session } = useAuth();

  return (
    <div style={styles.container}>
      <div style={styles.hero}>
        <h1 style={styles.title}>Welcome to the App</h1>
        <p style={styles.subtitle}>
          A modern application with secure authentication
        </p>

        <div style={styles.actions}>
          {session ? (
            <Link to="/dashboard" style={styles.primaryButton}>
              Go to Dashboard
            </Link>
          ) : (
            <>
              <Link to="/auth/login" style={styles.primaryButton}>
                Sign In
              </Link>
              <Link to="/auth/register" style={styles.secondaryButton}>
                Create Account
              </Link>
            </>
          )}
        </div>
      </div>

      <div style={styles.features}>
        <div style={styles.feature}>
          <h3>Secure Authentication</h3>
          <p>Built with Ory Kratos for enterprise-grade security</p>
        </div>
        <div style={styles.feature}>
          <h3>Account Management</h3>
          <p>Create and manage accounts with ease</p>
        </div>
        <div style={styles.feature}>
          <h3>Modern Stack</h3>
          <p>React, gRPC, and Kubernetes-native</p>
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
  hero: {
    backgroundColor: '#007bff',
    color: 'white',
    padding: '4rem 2rem',
    textAlign: 'center',
  },
  title: {
    fontSize: '2.5rem',
    marginBottom: '1rem',
  },
  subtitle: {
    fontSize: '1.25rem',
    opacity: 0.9,
    marginBottom: '2rem',
  },
  actions: {
    display: 'flex',
    gap: '1rem',
    justifyContent: 'center',
    flexWrap: 'wrap',
  },
  primaryButton: {
    padding: '0.75rem 2rem',
    backgroundColor: 'white',
    color: '#007bff',
    border: 'none',
    borderRadius: '4px',
    fontSize: '1rem',
    fontWeight: 600,
    cursor: 'pointer',
    textDecoration: 'none',
  },
  secondaryButton: {
    padding: '0.75rem 2rem',
    backgroundColor: 'transparent',
    color: 'white',
    border: '2px solid white',
    borderRadius: '4px',
    fontSize: '1rem',
    fontWeight: 600,
    cursor: 'pointer',
    textDecoration: 'none',
  },
  features: {
    display: 'flex',
    justifyContent: 'center',
    gap: '2rem',
    padding: '4rem 2rem',
    flexWrap: 'wrap',
  },
  feature: {
    backgroundColor: 'white',
    padding: '2rem',
    borderRadius: '8px',
    boxShadow: '0 2px 10px rgba(0, 0, 0, 0.1)',
    maxWidth: '300px',
    textAlign: 'center',
  },
};
