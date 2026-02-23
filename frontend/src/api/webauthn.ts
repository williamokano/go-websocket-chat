import type { AuthResponse, PasskeyInfo } from "../types";

// --- Base64url helpers ---

function base64urlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function base64urlDecode(str: string): ArrayBuffer {
  // Restore base64 padding
  let base64 = str.replace(/-/g, "+").replace(/_/g, "/");
  while (base64.length % 4 !== 0) {
    base64 += "=";
  }
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes.buffer;
}

// --- Registration ---

export async function beginRegistration(
  username: string
): Promise<{ sessionId: string; options: PublicKeyCredentialCreationOptions }> {
  const res = await fetch("/api/auth/webauthn/register/begin", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to begin registration");
  }
  const data = await res.json();
  const options = data.options;

  // Convert base64url strings to ArrayBuffers
  options.challenge = base64urlDecode(options.challenge);
  options.user.id = base64urlDecode(options.user.id);
  if (options.excludeCredentials) {
    options.excludeCredentials = options.excludeCredentials.map(
      (cred: { id: string; type: string; transports?: string[] }) => ({
        ...cred,
        id: base64urlDecode(cred.id),
      })
    );
  }

  return { sessionId: data.session_id, options };
}

export async function finishRegistration(
  sessionId: string,
  credential: PublicKeyCredential
): Promise<AuthResponse> {
  const response = credential.response as AuthenticatorAttestationResponse;
  const body = {
    id: credential.id,
    rawId: base64urlEncode(credential.rawId),
    type: credential.type,
    response: {
      attestationObject: base64urlEncode(response.attestationObject),
      clientDataJSON: base64urlEncode(response.clientDataJSON),
    },
  };

  const res = await fetch(
    `/api/auth/webauthn/register/finish?session_id=${encodeURIComponent(sessionId)}`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }
  );
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || "Failed to finish registration");
  }
  return res.json();
}

// --- Login ---

export async function beginLogin(): Promise<{
  sessionId: string;
  options: PublicKeyCredentialRequestOptions;
}> {
  const res = await fetch("/api/auth/webauthn/login/begin", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to begin login");
  }
  const data = await res.json();
  const options = data.options;

  // Convert base64url strings to ArrayBuffers
  options.challenge = base64urlDecode(options.challenge);
  if (options.allowCredentials) {
    options.allowCredentials = options.allowCredentials.map(
      (cred: { id: string; type: string; transports?: string[] }) => ({
        ...cred,
        id: base64urlDecode(cred.id),
      })
    );
  }

  return { sessionId: data.session_id, options };
}

export async function finishLogin(
  sessionId: string,
  credential: PublicKeyCredential
): Promise<AuthResponse> {
  const response = credential.response as AuthenticatorAssertionResponse;
  const body = {
    id: credential.id,
    rawId: base64urlEncode(credential.rawId),
    type: credential.type,
    response: {
      authenticatorData: base64urlEncode(response.authenticatorData),
      clientDataJSON: base64urlEncode(response.clientDataJSON),
      signature: base64urlEncode(response.signature),
      userHandle: response.userHandle
        ? base64urlEncode(response.userHandle)
        : undefined,
    },
  };

  const res = await fetch(
    `/api/auth/webauthn/login/finish?session_id=${encodeURIComponent(sessionId)}`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }
  );
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || "Failed to finish login");
  }
  return res.json();
}

// --- Credential management (authenticated) ---

export async function beginAddCredential(
  token: string
): Promise<{ sessionId: string; options: PublicKeyCredentialCreationOptions }> {
  const res = await fetch("/api/auth/webauthn/credentials/begin", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to begin adding credential");
  }
  const data = await res.json();
  const options = data.options;

  options.challenge = base64urlDecode(options.challenge);
  options.user.id = base64urlDecode(options.user.id);
  if (options.excludeCredentials) {
    options.excludeCredentials = options.excludeCredentials.map(
      (cred: { id: string; type: string; transports?: string[] }) => ({
        ...cred,
        id: base64urlDecode(cred.id),
      })
    );
  }

  return { sessionId: data.session_id, options };
}

export async function finishAddCredential(
  token: string,
  sessionId: string,
  credential: PublicKeyCredential,
  friendlyName: string
): Promise<PasskeyInfo> {
  const response = credential.response as AuthenticatorAttestationResponse;
  const body = {
    id: credential.id,
    rawId: base64urlEncode(credential.rawId),
    type: credential.type,
    response: {
      attestationObject: base64urlEncode(response.attestationObject),
      clientDataJSON: base64urlEncode(response.clientDataJSON),
    },
  };

  const params = new URLSearchParams({
    session_id: sessionId,
    friendly_name: friendlyName,
  });
  const res = await fetch(
    `/api/auth/webauthn/credentials/finish?${params.toString()}`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(body),
    }
  );
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || "Failed to finish adding credential");
  }
  return res.json();
}

export async function listCredentials(
  token: string
): Promise<PasskeyInfo[]> {
  const res = await fetch("/api/auth/webauthn/credentials", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to list credentials");
  }
  return res.json();
}

export async function deleteCredential(
  token: string,
  id: string
): Promise<void> {
  const res = await fetch(
    `/api/auth/webauthn/credentials/${encodeURIComponent(id)}`,
    {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to delete credential");
  }
}
