import { TranslateForm } from '../components/TranslateForm';

export function Home() {
  return (
    <div className="min-h-screen flex flex-col">
      <main className="flex-1 container mx-auto px-4 py-8">
        <div className="mb-8 text-center">
          <h2 className="text-3xl font-bold mb-2">Text Translation</h2>
          <p className="text-muted-foreground">
            Translate text between multiple languages instantly and for free
          </p>
        </div>
        <TranslateForm />
      </main>
    </div>
  );
}
