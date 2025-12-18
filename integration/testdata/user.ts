// TODO:[#db-migration] migrate user table

const cache: Record<string, User> = {}

// TODO:[#cache-user] add cache layer for user reads
// depends-on: #db-migration
export async function getUser(id: string): Promise<User> {
  const cached = cache[id]
  if (cached) return cached

  const user = await fetchUserFromDB(id)
  cache[id] = user
  return user
}

async function fetchUserFromDB(id: string): Promise<User> {
  return { id, name: "Ada", email: "ada@example.com" }
}

type User = {
  id: string
  name: string
  email: string
}
