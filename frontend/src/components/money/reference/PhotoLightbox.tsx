import { useEffect, useMemo, useState } from 'react';
import { Edit3, Loader2, Save, Trash2, X } from 'lucide-react';
import { moneyApi } from '../../../services/money';
import { MoneyCragNode, MoneyUpload } from '../../../types/money';
import { cragProblems } from './model';
import { T } from './theme';
import { TopoOverlayEditor, TopoOverlaySvg } from './TopoOverlayEditor';
import { TopoOverlay, overlaysForUpload } from './topoOverlay';

export interface PhotoLightboxItem {
  upload: MoneyUpload;
  title?: string;
  contextLabel?: string;
  noteBody?: string;
  activeProblemId?: string | null;
  startInTopoEdit?: boolean;
}

export function PhotoLightbox({ item, root = null, boulder = null, canDelete = false, canWrite = false, onSaveTopo, onUpdateMetadata, onDelete, onClose }: { item: PhotoLightboxItem; root?: MoneyCragNode | null; boulder?: MoneyCragNode | null; canDelete?: boolean; canWrite?: boolean; onSaveTopo?: (problem: MoneyCragNode, overlay: TopoOverlay) => void; onUpdateMetadata?: (upload: MoneyUpload, metadata: { title: string | null; comments: string | null }) => Promise<MoneyUpload | void> | MoneyUpload | void; onDelete?: (upload: MoneyUpload) => void; onClose: () => void }) {
  const [upload, setUpload] = useState(item.upload);
  const { src, loading } = useUploadImageUrl(upload.id);
  const [topoControlsTarget, setTopoControlsTarget] = useState<HTMLDivElement | null>(null);
  const title = upload.title || item.title || upload.original_filename;
  const meta = photoMeta(upload);
  const overlays = useMemo(() => overlaysForUpload(root, upload.id), [root, upload.id]);
  const [topoEditMode, setTopoEditMode] = useState(Boolean(item.startInTopoEdit));
  const [editing, setEditing] = useState(false);
  const [titleDraft, setTitleDraft] = useState(title);
  const [commentsDraft, setCommentsDraft] = useState(upload.comments ?? '');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const canDrawTopo = Boolean(canWrite && root && (boulder ? cragProblems(boulder).length > 0 : overlays.length > 0 || item.activeProblemId));

  useEffect(() => {
    setUpload(item.upload);
  }, [item.upload]);

  useEffect(() => {
    setTitleDraft(title);
    setCommentsDraft(upload.comments ?? '');
    setError(null);
    setEditing(false);
  }, [title, upload.comments, upload.id]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !editing) onClose();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [editing, onClose]);

  const cancelEdit = () => {
    setTitleDraft(title);
    setCommentsDraft(upload.comments ?? '');
    setError(null);
    setEditing(false);
  };

  const saveMetadata = async () => {
    if (!onUpdateMetadata) return;
    const nextTitle = titleDraft.trim() || null;
    const nextComments = commentsDraft.trim() || null;
    setSaving(true);
    setError(null);
    try {
      const updated = await onUpdateMetadata(upload, { title: nextTitle, comments: nextComments });
      setUpload(updated || { ...upload, title: nextTitle ?? undefined, comments: nextComments ?? undefined });
      setEditing(false);
    } catch {
      setError('Could not save photo details.');
    } finally {
      setSaving(false);
    }
  };

  return <div role="presentation" onClick={onClose} style={{ position: 'fixed', inset: 0, zIndex: 90, background: 'rgba(8,5,4,0.82)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 18 }}>
    <div role="dialog" aria-modal="true" aria-label={title} onClick={event => event.stopPropagation()} style={{ width: 'min(1280px, 100%)', height: '88vh', maxHeight: 900, display: 'grid', gridTemplateRows: 'auto minmax(0,1fr)', background: T.surf, border: `1px solid ${T.line2}`, borderRadius: 16, boxShadow: T.shadow, overflow: 'hidden' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', borderBottom: `1px solid ${T.line}` }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 15, fontWeight: 800, color: T.ink, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{title}</div>
          {item.contextLabel && <div style={{ fontFamily: T.mono, fontSize: 11, color: T.mut, marginTop: 3 }}>{item.contextLabel}</div>}
        </div>
        {canDrawTopo && !topoEditMode && <button type="button" onClick={() => setTopoEditMode(true)} style={{ border: `1px solid ${T.accentDim}`, background: T.accentSoft, color: T.accent, borderRadius: 9, minHeight: 38, padding: '0 12px', display: 'flex', alignItems: 'center', gap: 7, cursor: 'pointer', fontWeight: 800 }}><Edit3 size={16} />Draw topo lines</button>}
        {canDelete && <button type="button" aria-label="Delete photo" onClick={() => { if (window.confirm('Delete this photo?')) onDelete?.(upload); }} style={{ border: `1px solid #8F4E45`, background: 'rgba(143,78,69,0.14)', color: '#E6A299', borderRadius: 9, height: 38, padding: '0 12px', display: 'flex', alignItems: 'center', gap: 7, cursor: 'pointer', fontWeight: 800 }}><Trash2 size={16} />Delete</button>}
        <button type="button" aria-label="Close photo" onClick={onClose} style={{ border: `1px solid ${T.line2}`, background: T.inset, color: T.ink, borderRadius: 9, width: 38, height: 38, display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}><X size={18} /></button>
      </div>
      <div className="money-photo-lightbox-content" style={{ minHeight: 0, display: 'grid', gridTemplateColumns: 'minmax(0,1fr) minmax(280px,340px)', background: T.inset, overflow: 'hidden' }}>
        <style>{lightboxCss}</style>
        <div className="money-photo-lightbox-stage" style={{ minHeight: 0, minWidth: 0, position: 'relative', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 14, overflow: 'auto' }}>
          {loading || !src ? <Loader2 size={28} color={T.accent} /> : root && topoEditMode ? <TopoOverlayEditor root={root} boulder={boulder} upload={upload} canWrite={canWrite} activeProblemId={item.activeProblemId} imageSrc={src} imageAlt={title} controlsPortalTarget={topoControlsTarget} onSave={(problem, overlay) => onSaveTopo?.(problem, overlay)} /> : <div style={{ position: 'relative', maxWidth: '100%', maxHeight: 'calc(88vh - 92px)', lineHeight: 0 }}><img src={src} alt={title} style={{ maxWidth: '100%', maxHeight: 'calc(88vh - 92px)', objectFit: 'contain', borderRadius: 10 }} /><TopoOverlaySvg overlays={overlays} activeProblemId={item.activeProblemId} /></div>}
        </div>
        <PhotoDetailsPanel upload={upload} title={title} contextLabel={item.contextLabel} noteBody={item.noteBody} meta={meta} canWrite={canWrite} editing={editing} titleDraft={titleDraft} commentsDraft={commentsDraft} saving={saving} error={error} topoEditMode={topoEditMode} onEdit={() => setEditing(true)} onTitleDraft={setTitleDraft} onCommentsDraft={setCommentsDraft} onSave={saveMetadata} onCancel={cancelEdit} onTopoControlsTarget={setTopoControlsTarget} />
      </div>
    </div>
  </div>;
}

export function UploadPhotoButton({ upload, ratio = '1 / 1', title, onOpen, canDelete = false, onDelete }: { upload: MoneyUpload; ratio?: string; title?: string; onOpen: () => void; canDelete?: boolean; onDelete?: (upload: MoneyUpload) => void }) {
  const { src, loading } = useUploadImageUrl(upload.id);
  const label = title || upload.original_filename;
  const baseStyle: React.CSSProperties = { width: '100%', aspectRatio: ratio, borderRadius: 8, border: `1px solid ${T.line}`, overflow: 'hidden' };

  if (loading || !src) return <div style={{ ...baseStyle, background: T.map.slot, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Loader2 size={16} color={T.accent} /></div>;

  return <div style={{ position: 'relative' }}>
    <button type="button" aria-label={`Open photo ${label}`} onClick={onOpen} style={{ ...baseStyle, display: 'block', padding: 0, background: 'transparent', cursor: 'zoom-in' }}>
      <img src={src} alt={label} style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
    </button>
    {canDelete && <button type="button" aria-label={`Delete photo ${label}`} onClick={event => { event.stopPropagation(); if (window.confirm('Delete this photo?')) onDelete?.(upload); }} style={{ position: 'absolute', top: 6, right: 6, width: 28, height: 28, borderRadius: 8, border: `1px solid #8F4E45`, background: 'rgba(20,10,8,0.82)', color: '#E6A299', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}><Trash2 size={14} /></button>}
  </div>;
}

export function useUploadImageUrl(uploadId: string) {
  const [src, setSrc] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let objectUrl: string | null = null;
    let cancelled = false;
    setSrc(null);
    if (!uploadId) {
      setLoading(false);
      return () => {
        cancelled = true;
      };
    }
    setLoading(true);
    moneyApi.getUploadBlobUrl(uploadId).then(url => {
      if (cancelled) {
        if (url.startsWith('blob:')) URL.revokeObjectURL(url);
        return;
      }
      objectUrl = url.startsWith('blob:') ? url : null;
      setSrc(url);
    }).catch(() => {
      if (!cancelled) setSrc(null);
    }).finally(() => {
      if (!cancelled) setLoading(false);
    });
    return () => {
      cancelled = true;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [uploadId]);

  return { src, loading };
}

const lightboxCss = `
@media (max-width: 760px) {
  .money-photo-lightbox-content { grid-template-columns: minmax(0, 1fr) !important; overflow-y: auto !important; }
  .money-photo-lightbox-stage { min-height: 42dvh !important; max-height: none !important; padding: 8px !important; overflow: visible !important; }
  .money-photo-details-panel { border-left: none !important; border-top: 1px solid rgba(238,225,211,0.14) !important; max-height: none !important; }
}
`;

function PhotoDetailsPanel({ upload, title, contextLabel, noteBody, meta, canWrite, editing, titleDraft, commentsDraft, saving, error, topoEditMode, onEdit, onTitleDraft, onCommentsDraft, onSave, onCancel, onTopoControlsTarget }: { upload: MoneyUpload; title: string; contextLabel?: string; noteBody?: string; meta: string[]; canWrite: boolean; editing: boolean; titleDraft: string; commentsDraft: string; saving: boolean; error: string | null; topoEditMode: boolean; onEdit: () => void; onTitleDraft: (value: string) => void; onCommentsDraft: (value: string) => void; onSave: () => void; onCancel: () => void; onTopoControlsTarget: (node: HTMLDivElement | null) => void }) {
  const dirty = titleDraft.trim() !== (upload.title || upload.original_filename) || commentsDraft.trim() !== (upload.comments ?? '');
  return <aside className="money-photo-details-panel" style={{ minHeight: 0, maxHeight: 'calc(88vh - 58px)', overflowY: 'auto', padding: 16, borderLeft: `1px solid rgba(238,225,211,0.14)`, background: T.surf }}>
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
      <InfoLabel>Photo details</InfoLabel>
      {canWrite && !editing && <button type="button" aria-label="Edit photo details" onClick={onEdit} style={{ marginLeft: 'auto', border: `1px solid ${T.line2}`, background: T.inset, color: T.ink, borderRadius: 8, minHeight: 34, padding: '0 10px', display: 'flex', alignItems: 'center', gap: 6, cursor: 'pointer', fontWeight: 800 }}><Edit3 size={14} />Edit</button>}
    </div>
    {editing ? <div style={{ display: 'grid', gap: 10 }}>
      <label style={editLabel}>Title<input aria-label="Photo title" value={titleDraft} onChange={event => onTitleDraft(event.target.value)} maxLength={200} style={editField} /></label>
      <label style={editLabel}>Comments<textarea aria-label="Photo comments" value={commentsDraft} onChange={event => onCommentsDraft(event.target.value)} rows={6} maxLength={5000} style={{ ...editField, resize: 'vertical', lineHeight: 1.45 }} /></label>
      {error && <div role="alert" style={{ color: '#E6A299', fontSize: 12 }}>{error}</div>}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
        <button type="button" onClick={onSave} disabled={saving || !dirty} style={{ ...panelButton, background: dirty ? T.accent : T.inset, color: dirty ? T.onAccent : T.faint, borderColor: dirty ? T.accent : T.line2 }}><Save size={14} />{saving ? 'Saving…' : 'Save'}</button>
        <button type="button" onClick={onCancel} disabled={saving} style={panelButton}>Cancel</button>
      </div>
    </div> : <>
      <div style={{ fontSize: 17, fontWeight: 900, color: T.ink, overflowWrap: 'anywhere', marginBottom: 14 }}>{title}</div>
      <InfoLabel>Comments</InfoLabel>
      <p style={{ margin: '0 0 16px', minHeight: 22, color: upload.comments ? '#D9CBBD' : T.faint, fontSize: 13, lineHeight: 1.5, whiteSpace: 'pre-wrap' }}>{upload.comments || 'No photo comments yet.'}</p>
    </>}
    {contextLabel && <><InfoLabel>From</InfoLabel><div style={{ fontSize: 12, color: '#D9CBBD', lineHeight: 1.4, marginBottom: 12 }}>{contextLabel}</div></>}
    {noteBody && <><InfoLabel>Associated note</InfoLabel><p style={{ margin: '0 0 12px', fontSize: 12, lineHeight: 1.45, color: '#D9CBBD', whiteSpace: 'pre-wrap' }}>{noteBody}</p></>}
    {meta.length > 0 && <><InfoLabel>Metadata</InfoLabel><div style={{ display: 'grid', gap: 5 }}>{meta.map(row => <div key={row} style={{ fontFamily: T.mono, fontSize: 10.5, color: T.mut }}>{row}</div>)}</div></>}
    <div className="money-topo-controls-slot" ref={onTopoControlsTarget} style={{ marginTop: 16, display: topoEditMode ? 'grid' : 'none', gap: 8 }} />
  </aside>;
}

function InfoLabel({ children }: { children: React.ReactNode }) {
  return <div style={{ fontFamily: T.mono, fontSize: 10.5, letterSpacing: 0.8, color: T.mut, textTransform: 'uppercase', margin: '0 0 6px' }}>{children}</div>;
}

const editLabel: React.CSSProperties = { display: 'grid', gap: 6, color: T.mut, fontFamily: T.mono, fontSize: 10.5, textTransform: 'uppercase', letterSpacing: 0.6 };
const editField: React.CSSProperties = { width: '100%', background: T.surf2, border: `1px solid ${T.line}`, borderRadius: 8, padding: '10px 11px', color: T.ink, fontFamily: T.font, fontSize: 13, outline: 'none', textTransform: 'none', letterSpacing: 0 };
const panelButton: React.CSSProperties = { border: `1px solid ${T.line2}`, background: T.inset, color: T.ink, borderRadius: 9, minHeight: 40, padding: '0 10px', cursor: 'pointer', fontWeight: 800, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 6 };

function photoMeta(upload: MoneyUpload): string[] {
  const rows: string[] = [];
  if (upload.content_type) rows.push(upload.content_type);
  if (upload.width && upload.height) rows.push(`${upload.width} × ${upload.height}`);
  if (upload.created_at) rows.push(`Uploaded ${new Date(upload.created_at).toLocaleString()}`);
  return rows;
}
