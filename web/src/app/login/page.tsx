/**
 * Login page placeholder.
 * Will be replaced with Slack OAuth login flow.
 */
export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold mb-4">Progress Checker</h1>
        <p className="text-muted-foreground mb-6">Sign in to continue</p>
        <button
          type="button"
          className="rounded-md bg-primary px-6 py-2 text-primary-foreground hover:opacity-90"
        >
          Sign in with Slack
        </button>
      </div>
    </main>
  );
}
