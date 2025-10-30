import { getMe } from "./api/getMe.ts";

let myUserId: string | undefined = undefined;
export async function loadMe() {
  const me = await getMe();
  myUserId = me.id;
}

export function getMyUserId(): string {
  if (!myUserId) {
    throw new Error("My user ID not loaded yet. Call loadMe() first.");
  }
  return myUserId;
}
