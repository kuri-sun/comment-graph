// TODO: cleanup-sample remove legacy endpoints
// DEPS: #cache-sample

import { getUser } from "./users"

export async function handler(id: string) {
  const user = await getUser(id)
  return { status: 200, body: user }
}
