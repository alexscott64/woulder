import { ChangeEvent, useEffect, useState } from 'react';
import { Camera, CheckCircle2, FileImage, Image as ImageIcon, Loader2, Trash2, Upload } from 'lucide-react';
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

  if (failed) return <div className="flex aspect-video items-center justify-center rounded-2xl border border-[#2B403A] bg-[#0B1714]/80 text-xs font-semibold text-[#AAB8AD]">Image unavailable</div>;
  if (!src) return <div className="flex aspect-video items-center justify-center rounded-2xl border border-[#2B403A] bg-[#0B1714]/80"><Loader2 className="h-5 w-5 animate-spin text-[#7EA16B]" /></div>;
  return <img src={src} alt={upload.original_filename} width={upload.width} height={upload.height} loading="lazy" className="aspect-video w-full rounded-2xl border border-[#2B403A] object-cover" />;
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
    <section className="space-y-3 text-[#F2F0E7]">
      <div className="rounded-2xl border border-[#2B403A] bg-[#111D1B] p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-[#172522] text-[#7EA16B]">
              <ImageIcon className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#7EA16B]">Photos</p>
              <h3 className="text-lg font-semibold text-[#F2F0E7]">Photo references</h3>
              <p className="text-xs text-[#AAB8AD]">Photos tied to this selected map item</p>
            </div>
          </div>
          {canWrite && (
            <label className="flex cursor-pointer items-center gap-2 rounded-xl bg-[#7EA16B] px-4 py-3 text-sm font-semibold text-[#07110F] hover:bg-[#92B57D]">
              {uploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Upload className="h-4 w-4" />}
              {uploading ? 'Uploading' : 'Add photo'}
              <input type="file" accept="image/jpeg,image/png,image/webp" capture="environment" onChange={handleFileChange} disabled={uploading} className="sr-only" />
            </label>
          )}
        </div>
        <div className="mt-3 grid gap-2 text-xs text-[#AAB8AD] sm:grid-cols-3">
          <span className="flex items-center gap-1 rounded-full bg-[#172522] px-3 py-2"><CheckCircle2 className="h-3.5 w-3.5 text-[#7EA16B]" />{uploading ? 'Uploading' : 'Ready'}</span>
          <span className="flex items-center gap-1 rounded-full bg-[#172522] px-3 py-2"><FileImage className="h-3.5 w-3.5 text-[#6F9FB5]" />{uploads.length} attached</span>
          <span className="flex items-center gap-1 rounded-full bg-[#172522] px-3 py-2"><Camera className="h-3.5 w-3.5 text-[#C7D38A]" />Map item</span>
        </div>
      </div>

      {error && <div className="rounded-2xl border border-red-900/60 bg-red-950/40 px-3 py-2 text-xs font-semibold text-red-200">{error}</div>}

      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
        {uploads.map(upload => (
          <div key={upload.id} className="rounded-2xl border border-[#2B403A] bg-[#111D1B] p-2">
            <AuthImage upload={upload} />
            <div className="flex items-center justify-between gap-3 px-2 py-2 text-xs text-[#AAB8AD]">
              <span className="truncate"><Camera className="mr-1 inline h-3.5 w-3.5" />{upload.original_filename} · {formatBytes(upload.byte_size)}</span>
              {canWrite && (
                <button type="button" onClick={() => handleDelete(upload.id)} className="rounded-full border border-[#2B403A] p-1.5 text-[#AAB8AD] hover:border-[#C88A3D] hover:text-[#E0B36F]" title="Delete image">
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              )}
            </div>
          </div>
        ))}
        {uploads.length === 0 && <div className="rounded-2xl border border-dashed border-[#2B403A] bg-[#111D1B]/75 p-5 text-center text-xs text-[#AAB8AD]">No photos yet. Add reference shots, topo photos, or condition context.</div>}
      </div>
    </section>
  );
}
