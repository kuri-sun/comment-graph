// TODO: remove legacy endpoints
// depends-on: #db-migration

import { getUser } from "./user";

export async function handler(id: string) {
  const user = await getUser(id);
  return { status: 200, body: user };
}
