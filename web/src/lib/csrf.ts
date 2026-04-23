export async function fetchCsrfToken(): Promise<string> {
  const response = await fetch('/api/csrf', { method: 'GET', credentials: 'same-origin' });
  if (!response.ok) {
    throw new Error('failed to load csrf token');
  }

  const payload = (await response.json()) as { token?: string };
  if (!payload.token) {
    throw new Error('csrf token missing');
  }

  return payload.token;
}