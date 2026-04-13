import { useState, useRef, useCallback } from 'react';
import { Upload, FileText, Download, Loader2, CheckCircle2, XCircle, Clock } from 'lucide-react';
import { Button } from './ui/button';
import { Select } from './ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { LANGUAGES, MAX_FILE_SIZE, ALLOWED_FILE_TYPES } from '@/lib/constants';
import { api } from '@/lib/api';
import { useTaskPolling } from '@/hooks/useTaskPolling';

export function DocumentUpload() {
  const [file, setFile] = useState<File | null>(null);
  const [sourceLang, setSourceLang] = useState('auto');
  const [targetLang, setTargetLang] = useState('EN');
  const [taskId, setTaskId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [downloadUrl, setDownloadUrl] = useState<string | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Use the polling hook
  const { taskStatus, isPolling } = useTaskPolling({
    taskId: taskId || '',
    enabled: !!taskId,
    onCompleted: async () => {
      // Download the result file
      try {
        const blob = await api.downloadTaskResult(taskId!);
        const url = URL.createObjectURL(blob);
        setDownloadUrl(url);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to download result');
      }
    },
    onFailed: (errorMsg) => {
      setError(errorMsg);
    },
  });

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (selectedFile) processFile(selectedFile);
  };

  const processFile = (selectedFile: File) => {
    if (!ALLOWED_FILE_TYPES.includes(selectedFile.type)) {
      setError('Please select a valid document file (.docx or .txt)');
      setFile(null);
      return;
    }

    if (selectedFile.size > MAX_FILE_SIZE) {
      setError('File size exceeds 10MB limit');
      setFile(null);
      return;
    }

    setError(null);
    setFile(selectedFile);
    setDownloadUrl(null);
  };

  const handleTranslate = async () => {
    if (!file) return;

    setError(null);
    setDownloadUrl(null);

    try {
      // Create translation task
      const response = await api.createTranslationTask(file, {
        source_lang: sourceLang,
        target_lang: targetLang,
      });

      setTaskId(response.task_id);
      // Polling will start automatically via the hook
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create translation task');
    }
  };

  const handleClear = () => {
    setFile(null);
    setTaskId(null);
    setError(null);
    setDownloadUrl(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);

    const droppedFile = e.dataTransfer.files?.[0];
    if (droppedFile) processFile(droppedFile);
  }, []);

  const handleReset = () => {
    setTaskId(null);
    setError(null);
    setDownloadUrl(null);
  };

  const handleDownload = async () => {
    if (!taskId || !file) return;

    try {
      const blob = await api.downloadTaskResult(taskId);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `translated-${file.name}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to download result');
    }
  };

  return (
    <Card className="w-full max-w-4xl mx-auto">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="h-6 w-6" />
          Document Translation
        </CardTitle>
        <CardDescription>
          Upload documents (.docx, .txt) for translation (max 10MB)
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Source Language</label>
            <Select
              value={sourceLang}
              onChange={(e) => setSourceLang(e.target.value)}
              disabled={isPolling}
            >
              <option value="auto">Auto Detect</option>
              {LANGUAGES.map((lang) => (
                <option key={lang.code} value={lang.code}>
                  {lang.flag} {lang.name}
                </option>
              ))}
            </Select>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Target Language</label>
            <Select
              value={targetLang}
              onChange={(e) => setTargetLang(e.target.value)}
              disabled={isPolling}
            >
              {LANGUAGES.map((lang) => (
                <option key={lang.code} value={lang.code}>
                  {lang.flag} {lang.name}
                </option>
              ))}
            </Select>
          </div>
        </div>

        <div
          className={`border-2 border-dashed rounded-lg p-8 transition-colors ${
            isDragOver ? 'border-primary bg-primary/5' : 'border-border'
          }`}
          onDragOver={handleDragOver}
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
        >
          <div className="flex flex-col items-center justify-center space-y-4">
            <Upload className="h-12 w-12 text-muted-foreground" />
            <div className="text-center">
              <p className="text-sm font-medium">Click to upload or drag and drop</p>
              <p className="text-xs text-muted-foreground mt-1">
                Documents (.docx, .txt) up to 10MB
              </p>
            </div>
            <input
              ref={fileInputRef}
              type="file"
              accept=".docx,.txt"
              onChange={handleFileSelect}
              disabled={isPolling}
              className="hidden"
              id="file-upload"
            />
            <Button
              type="button"
              variant="outline"
              disabled={isPolling}
              onClick={() => fileInputRef.current?.click()}
            >
              <span>Select File</span>
            </Button>
          </div>

          {file && (
            <div className="mt-4 p-3 bg-secondary rounded-md">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <FileText className="h-4 w-4" />
                  <span className="text-sm font-medium">{file.name}</span>
                  <span className="text-xs text-muted-foreground">
                    ({(file.size / 1024 / 1024).toFixed(2)} MB)
                  </span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleClear}
                  disabled={isPolling}
                >
                  Remove
                </Button>
              </div>
            </div>
          )}
        </div>

        {error && (
          <div className="text-sm text-destructive bg-destructive/10 p-3 rounded-md flex items-center gap-2">
            <XCircle className="h-4 w-4" />
            {error}
          </div>
        )}

        {taskId && taskStatus && (
          <div className="space-y-3">
            {/* Task ID display */}
            <div className="text-xs text-muted-foreground bg-secondary p-2 rounded-md flex items-center gap-2">
              <Clock className="h-3 w-3" />
              Task ID: {taskId}
            </div>

            {/* Status indicator */}
            <div className="flex items-center gap-2">
              {taskStatus.status === 'pending' && (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Task is queued...
                </div>
              )}
              {taskStatus.status === 'processing' && (
                <div className="flex items-center gap-2 text-sm">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Processing...
                </div>
              )}
              {taskStatus.status === 'completed' && (
                <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
                  <CheckCircle2 className="h-4 w-4" />
                  Completed!
                </div>
              )}
              {taskStatus.status === 'failed' && (
                <div className="flex items-center gap-2 text-sm text-destructive">
                  <XCircle className="h-4 w-4" />
                  Failed: {taskStatus.error || 'Unknown error'}
                </div>
              )}
            </div>

            {/* Progress bar */}
            {taskStatus.status === 'processing' && (
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span>Translating document...</span>
                  <span className="text-muted-foreground">
                    {taskStatus.progress.toFixed(0)}%
                  </span>
                </div>
                <div className="h-2 bg-secondary rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary transition-all duration-300"
                    style={{ width: `${taskStatus.progress}%` }}
                  />
                </div>
                {/* Batch information */}
                {taskStatus.current_batch && taskStatus.total_batches && (
                  <div className="text-xs text-muted-foreground text-center">
                    Processing batch {taskStatus.current_batch} of {taskStatus.total_batches}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {downloadUrl && (
          <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-md">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Download className="h-4 w-4 text-green-600 dark:text-green-400" />
                <span className="text-sm font-medium">Translation complete!</span>
              </div>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" onClick={handleReset}>
                  Translate New
                </Button>
                <Button size="sm" onClick={handleDownload}>
                  Download
                </Button>
              </div>
            </div>
          </div>
        )}

        <Button
          onClick={handleTranslate}
          disabled={!file || isPolling}
          className="w-full"
          size="lg"
        >
          {isPolling ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Processing...
            </>
          ) : (
            <>
              <Upload className="mr-2 h-4 w-4" />
              Translate Document
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  );
}
