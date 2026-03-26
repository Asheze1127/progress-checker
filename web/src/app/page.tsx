import { redirect } from "next/navigation";

/**
 * Root page redirects to the login page.
 */
export default function RootPage() {
  redirect("/login");
}
