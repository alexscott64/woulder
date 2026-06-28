import { beforeEach, describe, expect, it, vi } from 'vitest';
import { moneyApi } from '../money';
import { authApiClient, authorizedFetch } from '../auth';

vi.mock('../auth', () => ({
  authApiClient: {
    get: vi.fn(),
    patch: vi.fn(),
  },
  authorizedFetch: vi.fn(),
}));

describe('moneyApi upload downloads', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns signed private object URLs without proxying through the API', async () => {
    vi.mocked(authApiClient.get).mockResolvedValue({ data: { url: 'https://example.invalid/signed', proxy_url: '/api/money/uploads/upload-1', expires_at: '2026-01-01T00:00:00Z' } });

    await expect(moneyApi.getUploadBlobUrl('upload-1')).resolves.toBe('https://example.invalid/signed');
    expect(authApiClient.get).toHaveBeenCalledWith('/money/uploads/upload-1/download-url');
    expect(authorizedFetch).not.toHaveBeenCalled();
  });

  it('falls back to the authenticated proxy URL when no signed URL is available', async () => {
    const blob = new Blob(['hello'], { type: 'text/plain' });
    const objectURL = 'blob:money-upload';
    vi.spyOn(URL, 'createObjectURL').mockReturnValue(objectURL);
    vi.mocked(authApiClient.get).mockResolvedValue({ data: { url: '/api/money/uploads/upload-1', proxy_url: '/api/money/uploads/upload-1', expires_at: '2026-01-01T00:00:00Z' } });
    vi.mocked(authorizedFetch).mockResolvedValue({ ok: true, blob: () => Promise.resolve(blob) } as Response);

    await expect(moneyApi.getUploadBlobUrl('upload-1')).resolves.toBe(objectURL);
    expect(vi.mocked(authorizedFetch).mock.calls[0][0]).toMatch(/\/api\/money\/uploads\/upload-1$/);
  });
  it('updates editable upload metadata', async () => {
    vi.mocked(authApiClient.patch).mockResolvedValue({ data: { id: 'upload-1', title: 'Topo overview', comments: 'Main face' } });

    await expect(moneyApi.updateUploadMetadata('upload-1', { title: 'Topo overview', comments: 'Main face' })).resolves.toEqual(expect.objectContaining({ title: 'Topo overview', comments: 'Main face' }));
    expect(authApiClient.patch).toHaveBeenCalledWith('/money/uploads/upload-1/metadata', { title: 'Topo overview', comments: 'Main face' });
  });
});
