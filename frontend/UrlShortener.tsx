"use client";

import { useState } from "react";

type Props = {
  /** Base URL of the shortener API, e.g. "https://api.example.com". */
  apiBase?: string;
};

const EXPIRY_OPTIONS = [
  { label: "Never expires", value: 0 },
  { label: "Expires in 1 hour", value: 3600 },
  { label: "Expires in 1 day", value: 86400 },
  { label: "Expires in 1 week", value: 604800 },
];

export default function UrlShortener({
  apiBase = process.env.NEXT_PUBLIC_SHORTENER_API ?? "http://localhost:8080",
}: Props) {
  const [longUrl, setLongUrl] = useState("");
  const [alias, setAlias] = useState("");
  const [expiresIn, setExpiresIn] = useState(0);
  const [shortUrl, setShortUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setShortUrl(null);
    setCopied(false);

    try {
      const res = await fetch(`${apiBase}/shorten`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          long_url: longUrl,
          alias: alias.trim() || undefined,
          expires_in: expiresIn,
        }),
      });

      if (!res.ok) {
        const body = await res.text();
        throw new Error(
          body.replace(/^Error:\s*/i, "").trim() || `Request failed (${res.status})`,
        );
      }

      const data = (await res.json()) as { short_url: string };
      setShortUrl(data.short_url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  async function copy() {
    if (!shortUrl) return;
    await navigator.clipboard.writeText(shortUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <div className="us">
      <form onSubmit={onSubmit} className="us-form">
        <input
          type="url"
          required
          placeholder="https://example.com/a/very/long/link"
          value={longUrl}
          onChange={(e) => setLongUrl(e.target.value)}
          className="us-input"
        />
        <div className="us-row">
          <input
            type="text"
            placeholder="custom alias (optional)"
            pattern="[A-Za-z0-9_-]{3,16}"
            title="3–16 letters, digits, - or _"
            value={alias}
            onChange={(e) => setAlias(e.target.value)}
            className="us-input"
          />
          <select
            value={expiresIn}
            onChange={(e) => setExpiresIn(Number(e.target.value))}
            className="us-input"
            aria-label="Expiration"
          >
            {EXPIRY_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </div>
        <button type="submit" disabled={loading} className="us-btn">
          {loading ? "Shortening…" : "Shorten"}
        </button>
      </form>

      {shortUrl && (
        <div className="us-result">
          <a
            href={shortUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="us-link"
          >
            {shortUrl}
          </a>
          <button type="button" onClick={copy} className="us-copy">
            {copied ? "Copied" : "Copy"}
          </button>
        </div>
      )}

      {error && <p className="us-error">{error}</p>}

      <style jsx>{`
        .us {
          --us-accent: #4f46e5;
          --us-fg: #111827;
          --us-muted: #6b7280;
          --us-border: #e5e7eb;
          --us-bg: transparent;
          --us-err: #b91c1c;
          width: 100%;
          max-width: 480px;
          color: var(--us-fg);
          font-family: inherit;
        }
        @media (prefers-color-scheme: dark) {
          .us {
            --us-accent: #818cf8;
            --us-fg: #e5e7eb;
            --us-muted: #9ca3af;
            --us-border: #374151;
            --us-err: #fca5a5;
          }
        }
        .us-form {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }
        .us-row {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 10px;
        }
        .us-input {
          width: 100%;
          padding: 11px 13px;
          font-size: 0.95rem;
          color: var(--us-fg);
          background: var(--us-bg);
          border: 1px solid var(--us-border);
          border-radius: 9px;
          outline: none;
          transition: border-color 0.15s, box-shadow 0.15s;
        }
        .us-input:focus {
          border-color: var(--us-accent);
          box-shadow: 0 0 0 3px color-mix(in srgb, var(--us-accent) 22%, transparent);
        }
        .us-btn {
          padding: 11px 14px;
          font-size: 0.95rem;
          font-weight: 600;
          color: #fff;
          background: var(--us-accent);
          border: none;
          border-radius: 9px;
          cursor: pointer;
          transition: opacity 0.15s;
        }
        .us-btn:disabled {
          opacity: 0.6;
          cursor: progress;
        }
        .us-result {
          display: flex;
          gap: 8px;
          align-items: center;
          margin-top: 14px;
          padding: 11px 13px;
          border: 1px solid var(--us-border);
          border-radius: 9px;
        }
        .us-link {
          flex: 1;
          font-family: ui-monospace, monospace;
          font-size: 0.9rem;
          color: var(--us-accent);
          text-decoration: none;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .us-copy {
          padding: 7px 12px;
          font-size: 0.85rem;
          font-weight: 600;
          color: var(--us-fg);
          background: transparent;
          border: 1px solid var(--us-border);
          border-radius: 7px;
          cursor: pointer;
        }
        .us-error {
          margin: 12px 0 0;
          font-size: 0.9rem;
          color: var(--us-err);
        }
      `}</style>
    </div>
  );
}
