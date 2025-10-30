import { callManifoldApi } from "./callManifoldApi.ts";
import { User, UserSchema } from "./getUser.ts";

export async function getMe(): Promise<User> {
  const sb = await callManifoldApi("GET", "v0/me");
  return UserSchema.parse(JSON.parse(sb));
}
