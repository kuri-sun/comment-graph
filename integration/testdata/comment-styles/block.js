/*
 * @cgraph-id block-root
 * @cgraph-deps hash-root
 * Fetches data for the HTML shell.
 */
export async function loadDashboard() {
  return fetch("/api/dashboard").then((res) => res.json());
}
