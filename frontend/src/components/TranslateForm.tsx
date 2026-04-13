import { useState } from 'react';
import { Languages, ArrowRightLeft } from 'lucide-react';
import { Button } from './ui/button';
import { Textarea } from './ui/textarea';
import { Select } from './ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { LANGUAGES } from '@/lib/constants';
import { api } from '@/lib/api';

export function TranslateForm() {
  const [sourceText, setSourceText] = useState('');
  const [translatedText, setTranslatedText] = useState('');
  const [sourceLang, setSourceLang] = useState('auto');
  const [targetLang, setTargetLang] = useState('EN');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleTranslate = async () => {
    if (!sourceText.trim()) {
      setError('Please enter some text to translate');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const result = await api.translate({
        text: sourceText,
        source_lang: sourceLang,
        target_lang: targetLang,
      });
      setTranslatedText(result.data?.result || '');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Translation failed');
      setTranslatedText('');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSwap = () => {
    if (sourceLang !== 'auto') {
      setSourceLang(targetLang);
      setTargetLang(sourceLang);
    }
    if (translatedText) {
      setSourceText(translatedText);
      setTranslatedText(sourceText);
    }
  };

  const handleClear = () => {
    setSourceText('');
    setTranslatedText('');
    setError(null);
  };

  return (
    <Card className="w-full max-w-4xl mx-auto">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Languages className="h-6 w-6" />
          Text Translation
        </CardTitle>
        <CardDescription>
          Translate text between multiple languages for free
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-[1fr,auto,1fr] gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">From</label>
            <Select
              value={sourceLang}
              onChange={(e) => setSourceLang(e.target.value)}
              disabled={isLoading}
            >
              <option value="auto">Auto Detect</option>
              {LANGUAGES.map((lang) => (
                <option key={lang.code} value={lang.code}>
                  {lang.flag} {lang.name}
                </option>
              ))}
            </Select>
          </div>

          <div className="flex items-end">
            <Button
              variant="outline"
              size="icon"
              onClick={handleSwap}
              disabled={isLoading || sourceLang === 'auto'}
              title="Swap languages"
            >
              <ArrowRightLeft className="h-4 w-4" />
            </Button>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">To</label>
            <Select
              value={targetLang}
              onChange={(e) => setTargetLang(e.target.value)}
              disabled={isLoading}
            >
              {LANGUAGES.map((lang) => (
                <option key={lang.code} value={lang.code}>
                  {lang.flag} {lang.name}
                </option>
              ))}
            </Select>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Source Text</label>
            <Textarea
              value={sourceText}
              onChange={(e) => setSourceText(e.target.value)}
              placeholder="Enter text to translate..."
              className="min-h-[200px]"
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Translation</label>
            <Textarea
              value={translatedText}
              readOnly
              placeholder="Translation will appear here..."
              className="min-h-[200px]"
              disabled={isLoading}
            />
          </div>
        </div>

        {error && (
          <div className="text-sm text-destructive bg-destructive/10 p-3 rounded-md">
            {error}
          </div>
        )}

        <div className="flex gap-2">
          <Button
            onClick={handleTranslate}
            disabled={isLoading || !sourceText.trim()}
            className="flex-1"
          >
            {isLoading ? 'Translating...' : 'Translate'}
          </Button>
          <Button
            variant="outline"
            onClick={handleClear}
            disabled={isLoading}
          >
            Clear
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
