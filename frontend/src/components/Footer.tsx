export function Footer() {
  return (
    <footer className="border-t mt-auto">
      <div className="container mx-auto px-4 py-6">
        <div className="flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-sm text-muted-foreground">
            Powered by DeepLX - Free and open source translation
          </p>
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <a
              href="https://github.com/OwO-Network/DeepLX"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-foreground transition-colors"
            >
              DeepLX on GitHub
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
