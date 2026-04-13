import { DocumentUpload } from '../components/DocumentUpload';

export function Document() {
  return (
    <div className="min-h-screen flex flex-col">
      <main className="flex-1 container mx-auto px-4 py-8">
        <div className="mb-8 text-center">
          <h2 className="text-3xl font-bold mb-2">Document Translation</h2>
          <p className="text-muted-foreground">
            Upload Word documents for translation while preserving formatting
          </p>
        </div>
        <DocumentUpload />
      </main>
    </div>
  );
}
