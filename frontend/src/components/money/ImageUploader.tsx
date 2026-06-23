import { ChangeEvent, useEffect, useState } from 'react';
import { Camera, Image as ImageIcon, Loader2, Trash2, Upload } from 'lucide-react';
import { moneyApi } from '../../services/money';
import { MoneyUpload } from '../../types/money';
import { formatBytes } from './geometry';

interface ImageUploaderProps {
  projectId: string;
  featureId: string;
  uploads: MoneyUpload[];
  canWrite: boolean;
  onUploaded: () => void;
  onDeleted: () => void;
}

function AuthImage({ upload }: { upload: MoneyUpload }) {
  const [src, setSrc] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let url: string | null = null;
    moneyApi.getUploadBlobUrl(upload.id)
      .then(blobUrl => {
        url = blobUrl;
        setSrc(blobUrl);
      })
      .catch(() => setFailed(true));
    return () => {
      if (url) URL.revokeObjectURL(url);
    };
  }, [upload.id]);

  if (failed) {
    return <div className="flex aspect-video items-center justify-center rounded-2xl bg-white/10 text-xs text-slate-400">Image unavailable</div>;
  }

  if (!src) {
    return <div className="flex aspect-video items-center justify-center rounded-2xl bg-white/10"><Loader2 className="h-5 w-5 animate-spin text-slate-300" /></div>;
  }

  return <img src={src} alt={upload.original_filename} width={upload.width} height={upload.height} loading="lazy" className="aspect-video w-full rounded-2xl object-cover" />;
}

export function ImageUploader({ projectId, featureId, uploads, canWrite, onUploaded, onDeleted }: ImageUploaderProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    setUploading(true);
    setError(null);
    try {
      await moneyApi.uploadImage(projectId, file, { featureId });
      onUploaded();
    } catch {
      setError('Upload failed. Use JPEG, PNG, or WebP under the server limit.');
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (uploadId: string) => {
    setError(null);
    try {
      await moneyApi.deleteUpload(uploadId);
      onDeleted();
    } catch {
      setError('Could not delete image.');
    }
  };

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm font-bold text-white">
          <ImageIcon className="h-4 w-4 text-emerald-200" />
          Images
        </div>
        {canWrite && (
          <label className="flex cursor-pointer items-center gap-2 rounded-full bg-white/10 px-3 py-1.5 text-xs font-bold text-slate-100 hover:bg-white/15">
            {uploading ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Upload className="h-3.5 w-3.5" />}
            Add
            <input type="file" accept="image/jpeg,image/png,image/webp" capture="environment" onChange={handleFileChange} disabled={uploading} className="sr-only" />
          </label>
        )}
      </div>

      {error && <div className="rounded-2xl border border-red-300/20 bg-red-500/10 px-3 py-2 text-xs text-red-100">{error}</div>}

      <div className="grid gap-3">
        {uploads.map(upload => (
          <div key={upload.id} className="rounded-3xl border border-white/10 bg-white/8 p-2">
            <AuthImage upload={upload} />
            <div className="flex items-center justify-between gap-3 px-2 py-2 text-xs text-slate-400">
              <span className="truncate"><Camera className="mr-1 inline h-3.5 w-3.5" />{upload.original_filename} · {formatBytes(upload.byte_size)}</span>
              {canWrite && (
                <button onClick={() => handleDelete(upload.id)} className="rounded-full p-1.5 text-slate-300 hover:bg-red-400/20 hover:text-red-100" title="Delete image">
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              )}
            </div>
          </div>
        ))}
        {uploads.length === 0 && <div className="rounded-3xl border border-dashed border-white/15 p-5 text-center text-xs text-slate-400">No images yet.</div>}
      </div>
    </section>
  );
}
