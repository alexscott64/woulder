import { FormEvent, useState } from 'react';
import { Clock, Loader2, MessageSquare, Shield, Trash2, UserCircle2 } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import { moneyApi } from '../../services/money';
import { MoneyNote } from '../../types/money';

interface NotesPanelProps {
  featureId: string;
  notes: MoneyNote[];
  canWrite: boolean;
  onChanged: () => void;
}

function displayAuthor(note: MoneyNote, currentUserId?: string) {
  if (note.created_by === currentUserId) return 'You';
  return `Team member ${note.created_by.slice(0, 8)}`;
}

export function NotesPanel({ featureId, notes, canWrite, onChanged }: NotesPanelProps) {
  const { user } = useAuth();
  const [body, setBody] = useState('');
  const [visibility, setVisibility] = useState<'team' | 'private'>('team');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!body.trim()) return;
    setSaving(true);
    setError(null);
    try {
      await moneyApi.createNote(featureId, { body: body.trim(), visibility });
      setBody('');
      setVisibility('team');
      onChanged();
    } catch {
      setError('Could not save note.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (note: MoneyNote) => {
    setError(null);
    try {
      await moneyApi.deleteNote(note.id);
      onChanged();
    } catch {
      setError('Could not delete note.');
    }
  };

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between gap-3 rounded-2xl border border-[#2B403A] bg-[#111D1B] px-3 py-2">
        <div className="flex items-center gap-2 text-sm font-semibold text-[#F2F0E7]">
          <MessageSquare className="h-4 w-4 text-[#7EA16B]" />
          Notes
        </div>
        <span className="rounded-full bg-[#172522] px-2 py-1 text-xs font-semibold text-[#AAB8AD]">{notes.length}</span>
      </div>

      {canWrite && (
        <form onSubmit={handleSubmit} className="rounded-2xl border border-[#2B403A] bg-[#111D1B] p-3 text-[#F2F0E7]">
          <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.14em] text-[#7EA16B]">
            <UserCircle2 className="h-4 w-4" />
            Add note as {user?.display_name ?? 'current user'}
          </div>
          <textarea
            value={body}
            onChange={event => setBody(event.target.value)}
            className="min-h-28 w-full resize-none rounded-2xl border border-[#2B403A] bg-[#0B1714]/80 p-3 text-sm leading-6 text-[#F2F0E7] outline-none placeholder:text-[#74847B] focus:border-[#7EA16B] focus:ring-2 focus:ring-[#7EA16B]/10"
            placeholder="Beta, access detail, landing note, photo context, or next action..."
            maxLength={5000}
          />
          <div className="mt-2 flex items-center justify-between gap-2">
            <select value={visibility} onChange={event => setVisibility(event.target.value as 'team' | 'private')} className="rounded-xl border border-[#2B403A] bg-[#0B1714]/80 px-3 py-2 text-xs font-semibold text-[#F2F0E7]">
              <option value="team">Team visible</option>
              <option value="private">Private note</option>
            </select>
            <button type="submit" disabled={saving || !body.trim()} className="flex items-center gap-2 rounded-xl bg-[#7EA16B] px-4 py-2 text-xs font-semibold text-[#07110F] disabled:bg-[#2B403A] disabled:text-[#74847B]">
              {saving && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              Save note
            </button>
          </div>
        </form>
      )}

      {error && <div className="rounded-2xl border border-red-900/60 bg-red-950/40 px-3 py-2 text-xs font-semibold text-red-200">{error}</div>}

      <div className="space-y-3">
        {notes.map(note => {
          const canDelete = canWrite && (user?.id === note.created_by || user?.role === 'admin');
          return (
            <article key={note.id} className="rounded-2xl border border-[#2B403A] bg-[#111D1B] p-4 text-[#F2F0E7]">
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2 text-xs text-[#AAB8AD]">
                <span className="flex items-center gap-1"><UserCircle2 className="h-4 w-4 text-[#7EA16B]" />{displayAuthor(note, user?.id)}</span>
                <span className="flex items-center gap-1"><Shield className="h-3.5 w-3.5" />{note.visibility}</span>
                <span className="flex items-center gap-1"><Clock className="h-3.5 w-3.5" />{new Date(note.created_at).toLocaleString()}</span>
              </div>
              <p className="whitespace-pre-wrap rounded-2xl bg-[#0B1714]/80 p-3 text-sm leading-7 text-[#DDE6DC]">{note.body}</p>
              {canDelete && (
                <button type="button" onClick={() => handleDelete(note)} className="mt-3 inline-flex items-center gap-1 rounded-xl border border-[#2B403A] bg-[#111D1B] px-3 py-1.5 text-xs font-semibold text-[#AAB8AD] hover:border-[#C88A3D] hover:text-[#E0B36F]">
                  <Trash2 className="h-3.5 w-3.5" />
                  Delete
                </button>
              )}
            </article>
          );
        })}
        {notes.length === 0 && <div className="rounded-2xl border border-dashed border-[#2B403A] bg-[#111D1B]/75 p-5 text-center text-xs text-[#AAB8AD]">No notes yet. Add beta, access details, photo context, or a next action.</div>}
      </div>
    </section>
  );
}
