import { useCallback, useEffect, useState } from "react";
import type { PasskeyInfo } from "../../types";
import * as webauthnApi from "../../api/webauthn";

interface PasskeySettingsProps {
  token: string;
  onClose: () => void;
}

export function PasskeySettings({ token, onClose }: PasskeySettingsProps) {
  const [credentials, setCredentials] = useState<PasskeyInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [adding, setAdding] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [friendlyName, setFriendlyName] = useState("");
  const [showNameInput, setShowNameInput] = useState(false);

  const fetchCredentials = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const creds = await webauthnApi.listCredentials(token);
      setCredentials(creds);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load passkeys"
      );
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    fetchCredentials();
  }, [fetchCredentials]);

  const handleAdd = async () => {
    if (!friendlyName.trim()) {
      setError("Please enter a name for the passkey");
      return;
    }
    setError("");
    setAdding(true);
    try {
      const { sessionId, options } =
        await webauthnApi.beginAddCredential(token);
      const credential = (await navigator.credentials.create({
        publicKey: options,
      })) as PublicKeyCredential | null;
      if (!credential) {
        throw new Error("Passkey creation was cancelled");
      }
      await webauthnApi.finishAddCredential(
        token,
        sessionId,
        credential,
        friendlyName.trim()
      );
      setFriendlyName("");
      setShowNameInput(false);
      await fetchCredentials();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add passkey");
    } finally {
      setAdding(false);
    }
  };

  const handleDelete = async (id: string) => {
    setError("");
    setDeletingId(id);
    try {
      await webauthnApi.deleteCredential(token, id);
      setCredentials((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to delete passkey"
      );
    } finally {
      setDeletingId(null);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">
            Passkey Settings
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <svg
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth="1.5"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-md bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {loading ? (
          <div className="py-8 text-center text-sm text-gray-400">
            Loading passkeys...
          </div>
        ) : (
          <>
            {credentials.length === 0 ? (
              <p className="py-4 text-center text-sm text-gray-500">
                No passkeys registered yet.
              </p>
            ) : (
              <ul className="mb-4 divide-y divide-gray-100">
                {credentials.map((cred) => (
                  <li
                    key={cred.id}
                    className="flex items-center justify-between py-3"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-gray-900">
                        {cred.friendly_name}
                      </p>
                      <p className="text-xs text-gray-500">
                        Created {formatDate(cred.created_at)}
                        {cred.last_used_at && (
                          <> &middot; Last used {formatDate(cred.last_used_at)}</>
                        )}
                      </p>
                    </div>
                    <button
                      onClick={() => handleDelete(cred.id)}
                      disabled={deletingId === cred.id}
                      className="ml-3 text-xs text-red-500 hover:text-red-700 disabled:opacity-50"
                    >
                      {deletingId === cred.id ? "Deleting..." : "Delete"}
                    </button>
                  </li>
                ))}
              </ul>
            )}

            {showNameInput ? (
              <div className="space-y-3">
                <div>
                  <label
                    htmlFor="passkey-name"
                    className="block text-sm font-medium text-gray-700 mb-1"
                  >
                    Passkey name
                  </label>
                  <input
                    id="passkey-name"
                    type="text"
                    value={friendlyName}
                    onChange={(e) => setFriendlyName(e.target.value)}
                    placeholder="e.g. MacBook Touch ID"
                    className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    autoFocus
                  />
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={handleAdd}
                    disabled={adding}
                    className="flex-1 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
                  >
                    {adding ? "Adding..." : "Add Passkey"}
                  </button>
                  <button
                    onClick={() => {
                      setShowNameInput(false);
                      setFriendlyName("");
                    }}
                    disabled={adding}
                    className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setShowNameInput(true)}
                className="w-full rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              >
                Add Passkey
              </button>
            )}
          </>
        )}
      </div>
    </div>
  );
}
