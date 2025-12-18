// TODO:[#cleanup-legacy] remove legacy endpoints
// depends-on: #cache-user

import { getUser } from "./user"

export async function handler(id: string) {
  const user = await getUser(id)
  return { status: 200, body: user }
}
