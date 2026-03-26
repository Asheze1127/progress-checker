/**
 * Layout wrapper for authenticated routes.
 * Will be extended with auth checks and navigation.
 */
export default function AuthenticatedLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="min-h-screen">
      <header className="border-b px-6 py-4">
        <nav className="flex items-center gap-6">
          <span className="font-bold">Progress Checker</span>
          <a href="/progress" className="text-sm text-muted-foreground hover:text-foreground">
            Progress
          </a>
          <a href="/questions" className="text-sm text-muted-foreground hover:text-foreground">
            Questions
          </a>
        </nav>
      </header>
      <main className="p-6">{children}</main>
    </div>
  );
}
