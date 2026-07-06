import { createTheme } from '@mantine/core';

export const theme = createTheme({
  cursorType: 'pointer',
  fontFamily: 'Montserrat, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  fontFamilyMonospace: 'Fira Mono, monospace',
  scale: 1,
  fontSmoothing: true,
  focusRing: 'never',
  white: '#ffffff',
  black: '#24292f',
  colors: {
    dark: [
      '#c9d1d9',
      '#b1bac4',
      '#8b949e',
      '#6e7681',
      '#484f58',
      '#30363d',
      '#21262d',
      '#161b22',
      '#0d1117',
      '#010409',
    ],
  },
  primaryShade: 8,
  primaryColor: 'cyan',
  autoContrast: true,
  luminanceThreshold: 0.3,
  headings: {
    fontWeight: '600',
  },
  defaultRadius: 'md',
});
