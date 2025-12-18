// TODO: cleanup-sample remove legacy endpoints
// depends-on: #cache-sample

import { getUser } from "./users"

export async function handler(id: string) {
  const user = await getUser(id)
  return { status: 200, body: user }
}
