// Used for __tests__/testing-library.js
// Learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom/extend-expect';

import { render } from '@testing-library/react';
import App from './App';
import { it, describe, expect } from 'vitest';

/**
* @vitest-environment jsdom
*/
describe('app', () => {
  it('renders without crashing', () => {
    const { container } = render(<App />);
    expect(container).toBeInTheDocument();
  })
});
