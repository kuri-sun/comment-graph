// TODO:[#db-migration] migrate user table

// TODO:[#cache-user] add cache layer for user reads
// depends-on: #db-migration
export async function getUser(id: string) {
  const user = await fetchUserFromDB(id)
  return user
}

async function fetchUserFromDB(id: string) {
  return { id, name: "Ada" }
}
