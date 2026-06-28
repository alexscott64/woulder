import { useEffect, useRef, useState } from 'react';
import { Camera, FileText, PenLine } from 'lucide-react';
import { MoneyCragNode, MoneyNote, MoneyNoteBlock, MoneyNoteTargetType, MoneyUpload } from '../../../types/money';
import { useUploadImageUrl } from './PhotoLightbox';
import { flattenAreas, flattenBoulders } from './model';
import { T } from './theme';
import { SketchPadDialog } from './SketchPadDialog';

export interface NoteComposerPayload {
  body: string;
  tags: string[];
  target_type: MoneyNoteTargetType;
  target_ref?: string;
  blocks: MoneyNoteBlock[];
  files: Array<{ file: File; kind: 'photo' | 'file'; blockKey: string }>;
}

interface Props {
  root: MoneyCragNode | null;
  area: MoneyCragNode;
  boulder: MoneyCragNode | null;
  initialBlock?: 'photo' | 'sketch' | 'file' | null;
  initialNote?: MoneyNote | null;
  uploads?: MoneyUpload[];
  mobile: boolean;
  onClose: () => void;
  onSubmit: (payload: NoteComposerPayload) => void;
}

type PendingFile = { file: File; kind: 'photo' | 'file'; blockKey: string };

export function NoteComposer({ root, area, boulder, initialBlock, initialNote, uploads = [], mobile, onClose, onSubmit }: Props) {
  const editing = Boolean(initialNote);
  const [body, setBody] = useState(initialNote?.body ?? '');
  const [tags, setTags] = useState((initialNote?.tags ?? []).join(', '));
  const [blocks, setBlocks] = useState<MoneyNoteBlock[]>(() => initialNote?.blocks ?? []);
  const [files, setFiles] = useState<PendingFile[]>([]);
  const [sketchOpen, setSketchOpen] = useState(initialBlock === 'sketch');
  const initialTargetType = initialNote?.target_type ?? (boulder ? 'boulder' : 'area');
  const initialTargetRef = initialNote?.target_ref ?? initialNote?.feature_id ?? (boulder ? boulder.feature.id : area.feature.id);
  const [target, setTarget] = useState(`${initialTargetType}:${initialTargetRef ?? ''}`);
  const areas = flattenAreas(root);
  const boulders = flattenBoulders(root);

  const objectUrls = useRef(new Set<string>());
  const uploadById = new Map(uploads.map(upload => [upload.id, upload]));
  const addFiles = (selectedFiles: File[], kind: 'photo' | 'file') => {
    if (!selectedFiles.length) return;
    const pendingFiles = selectedFiles.map(file => ({ file, kind, blockKey: `${Date.now()}-${Math.random().toString(36).slice(2)}` }));
    setFiles(current => [...current, ...pendingFiles]);
    setBlocks(current => [
      ...current,
      ...pendingFiles.map(({ file, blockKey }) => {
        const url = kind === 'photo' ? URL.createObjectURL(file) : undefined;
        if (url) objectUrls.current.add(url);
        return { kind, name: file.name, url, metadata: { local_block_key: blockKey } };
      }),
    ]);
  };
  const removeBlock = (index: number) => {
    const block = blocks[index];
    const localBlockKey = typeof block?.metadata?.local_block_key === 'string' ? block.metadata.local_block_key : null;
    if (block?.url?.startsWith('blob:')) {
      URL.revokeObjectURL(block.url);
      objectUrls.current.delete(block.url);
    }
    setBlocks(x => x.filter((_, i) => i !== index));
    if (localBlockKey) setFiles(x => x.filter(file => file.blockKey !== localBlockKey));
  };
  useEffect(() => () => {
    objectUrls.current.forEach(url => URL.revokeObjectURL(url));
    objectUrls.current.clear();
  }, []);

  const canPost = Boolean(body.trim() || blocks.length);
  const post = () => {
    const [type, id] = target.split(':');
    onSubmit({ body: body.trim(), tags: tags.split(',').map(s => s.trim()).filter(Boolean), target_type: type as MoneyNoteTargetType, target_ref: id || undefined, blocks, files });
  };

  return <div onClick={onClose} style={{ position: 'fixed', inset: 0, zIndex: 70, background: 'rgba(8,5,4,0.62)', display: 'flex', alignItems: mobile ? 'flex-end' : 'center', justifyContent: 'center' }}>
    <div onClick={e => e.stopPropagation()} style={{ background: T.surf, border: `1px solid ${T.line2}`, borderRadius: mobile ? '18px 18px 0 0' : 16, boxShadow: T.shadow, width: mobile ? '100%' : 460, maxHeight: '94vh', overflowY: 'auto' }}>
      <div style={{ padding: mobile ? '4px 18px 20px' : 22 }}>
        <div style={{ display: 'flex', alignItems: 'center', marginBottom: 14 }}>
          <span style={{ fontSize: 18, fontWeight: 800, color: T.ink }}>{editing ? 'Edit note' : 'New note'}</span>
          <button onClick={onClose} style={{ marginLeft: 'auto', border: 'none', background: 'transparent', color: T.mut, cursor: 'pointer', fontSize: 20 }}>×</button>
        </div>
        <Label>Link to</Label>
        <select value={target} onChange={e => setTarget(e.target.value)} style={inp}>
          <option value="none:">No link · general log</option>
          {areas.map(a => <option key={a.feature.id} value={`area:${a.feature.id}`}>{a.feature.title}</option>)}
          {boulders.map(b => <option key={b.feature.id} value={`boulder:${b.feature.id}`}>◈ {b.feature.title}</option>)}
        </select>
        <textarea value={body} onChange={e => setBody(e.target.value)} placeholder="What did you find? Beta, conditions, an idea…" rows={mobile ? 3 : 4} style={{ ...inp, resize: 'vertical', lineHeight: 1.5, marginTop: 12 }} />
        {blocks.length > 0 && <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, margin: '12px 0' }}>{blocks.map((bl, i) => <AttachmentPreview key={`${bl.upload_id ?? bl.name ?? bl.kind}-${i}`} block={bl} upload={bl.upload_id ? uploadById.get(bl.upload_id) : undefined} onRemove={() => removeBlock(i)} />)}</div>}
        <div style={{ display: 'flex', gap: 8, marginTop: 10 }}>
          <UploadButton icon={<Camera size={18} />} label="Photos" accept="image/*" onFiles={selectedFiles => addFiles(selectedFiles, 'photo')} />
          <ActionButton icon={<PenLine size={18} />} label="Sketch" onClick={() => setSketchOpen(true)} />
          <UploadButton icon={<FileText size={18} />} label="Files" onFiles={selectedFiles => addFiles(selectedFiles, 'file')} />
        </div>
        <input value={tags} onChange={e => setTags(e.target.value)} placeholder="tags, comma separated" style={{ ...inp, marginTop: 12, fontFamily: T.mono, fontSize: 12 }} />
        <div style={{ display: 'flex', gap: 8, marginTop: 14 }}>
          <button disabled={!canPost} onClick={post} style={{ flex: 1, border: 'none', borderRadius: 10, padding: 13, background: canPost ? T.accent : T.line2, color: canPost ? T.onAccent : T.faint, fontWeight: 800, cursor: canPost ? 'pointer' : 'default' }}>{editing ? 'Save note' : 'Post note'}</button>
          <button onClick={onClose} style={{ border: `1px solid ${T.line2}`, borderRadius: 10, padding: '13px 16px', background: 'transparent', color: T.ink, fontWeight: 700, cursor: 'pointer' }}>Cancel</button>
        </div>
      </div>
    </div>
    {sketchOpen && <SketchPadDialog onClose={() => setSketchOpen(false)} onSave={block => { setBlocks(current => [...current, block]); setSketchOpen(false); }} />}
  </div>;
}

function Label({ children }: { children: React.ReactNode }) { return <div style={{ fontFamily: T.mono, fontSize: 10.5, color: T.faint, textTransform: 'uppercase', letterSpacing: 0.5, marginBottom: 6 }}>{children}</div>; }
const inp: React.CSSProperties = { width: '100%', background: T.surf2, border: `1px solid ${T.line2}`, borderRadius: 9, padding: '10px 12px', color: T.ink, fontFamily: T.font, fontSize: 14, outline: 'none', boxSizing: 'border-box' };
const addBtn: React.CSSProperties = { flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 5, border: `1px solid ${T.line2}`, background: 'transparent', color: T.mut, borderRadius: 10, padding: '11px 6px', minHeight: 58, cursor: 'pointer', fontWeight: 700, fontSize: 11.5 };
function UploadButton({ icon, label, accept, onFiles }: { icon: React.ReactNode; label: string; accept?: string; onFiles: (files: File[]) => void }) { return <label style={addBtn}>{icon}{label}<input type="file" accept={accept} multiple style={{ display: 'none' }} onChange={e => { const selectedFiles = Array.from(e.currentTarget.files ?? []); onFiles(selectedFiles); e.currentTarget.value = ''; }} /></label>; }
function ActionButton({ icon, label, onClick }: { icon: React.ReactNode; label: string; onClick: () => void }) { return <button type="button" onClick={onClick} style={addBtn}>{icon}{label}</button>; }
function AttachmentPreview({ block, upload, onRemove }: { block: MoneyNoteBlock; upload?: MoneyUpload; onRemove: () => void }) { const isImage = block.kind === 'photo' || upload?.content_type.startsWith('image/'); const isVectorSketch = block.kind === 'sketch' && Boolean(block.metadata?.sketchpad); const { src } = useUploadImageUrl(upload && isImage ? upload.id : ''); const label = block.name ?? upload?.original_filename ?? block.kind; const imageSrc = block.url ?? src; return <div style={{ position: 'relative' }}>{isImage && imageSrc ? <img src={imageSrc} alt={label} style={{ width: 72, height: 72, objectFit: 'cover', borderRadius: 8, border: `1px solid ${T.line2}` }} /> : isVectorSketch ? <VectorSketchThumb block={block} label={label} /> : <div style={{ display: 'flex', alignItems: 'center', gap: 6, height: 72, padding: '0 12px', background: T.surf2, border: `1px solid ${T.line2}`, borderRadius: 8, color: T.ink, fontSize: 12, maxWidth: 150 }}><FileText size={16} />{label}</div>}<button aria-label={`Remove attachment ${label}`} onClick={onRemove} style={{ position: 'absolute', top: -7, right: -7, width: 22, height: 22, borderRadius: '50%', border: 'none', background: '#B65B4D', color: '#fff', cursor: 'pointer' }}>×</button></div>; }
function VectorSketchThumb({ block, label }: { block: MoneyNoteBlock; label: string }) { const sketchpad = block.metadata?.sketchpad; const strokes = Array.isArray(sketchpad && (sketchpad as { strokes?: unknown }).strokes) ? (sketchpad as { strokes: Array<{ id: string; points: Array<[number, number]>; color: string; width: number }> }).strokes : []; return <div aria-label={label} style={{ width: 72, height: 72, borderRadius: 8, border: `1px solid ${T.line2}`, background: '#fffaf0', position: 'relative', overflow: 'hidden' }}><svg viewBox="0 0 1000 1000" preserveAspectRatio="none" style={{ width: '100%', height: '100%' }}>{strokes.map(stroke => <path key={stroke.id} d={stroke.points.map((point, i) => `${i ? 'L' : 'M'} ${point[0] * 1000} ${point[1] * 1000}`).join(' ')} fill="none" stroke={stroke.color} strokeWidth={stroke.width * 2} strokeLinecap="round" strokeLinejoin="round" />)}</svg></div>; }
