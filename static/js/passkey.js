(function () {
	function toBase64Url(buffer) {
		const bytes = new Uint8Array(buffer);
		let binary = "";
		for (let i = 0; i < bytes.byteLength; i += 1) {
			binary += String.fromCharCode(bytes[i]);
		}
		return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
	}

	function fromBase64Url(base64url) {
		if (typeof base64url !== "string" || base64url.length === 0) {
			throw new Error("Invalid WebAuthn payload: expected base64url string");
		}
		const padded = base64url.replace(/-/g, "+").replace(/_/g, "/") + "===".slice((base64url.length + 3) % 4);
		const binary = atob(padded);
		const bytes = new Uint8Array(binary.length);
		for (let i = 0; i < binary.length; i += 1) {
			bytes[i] = binary.charCodeAt(i);
		}
		return bytes.buffer;
	}

	function normalizeCreationOptions(options) {
		const normalized = structuredClone(options);
		if (!normalized || !normalized.user) {
			throw new Error("Invalid registration options from server");
		}
		normalized.challenge = fromBase64Url(normalized.challenge);
		normalized.user.id = fromBase64Url(normalized.user.id);
		if (Array.isArray(normalized.excludeCredentials)) {
			normalized.excludeCredentials = normalized.excludeCredentials.map((credential) => ({
				...credential,
				id: fromBase64Url(credential.id),
			}));
		}
		return normalized;
	}

	function normalizeRequestOptions(options) {
		const normalized = structuredClone(options);
		if (!normalized) {
			throw new Error("Invalid login options from server");
		}
		normalized.challenge = fromBase64Url(normalized.challenge);
		if (Array.isArray(normalized.allowCredentials)) {
			normalized.allowCredentials = normalized.allowCredentials.map((credential) => ({
				...credential,
				id: fromBase64Url(credential.id),
			}));
		}
		return normalized;
	}

	function serializeCredential(credential) {
		return {
			id: credential.id,
			rawId: toBase64Url(credential.rawId),
			type: credential.type,
			response: {
				clientDataJSON: toBase64Url(credential.response.clientDataJSON),
				attestationObject: credential.response.attestationObject
					? toBase64Url(credential.response.attestationObject)
					: undefined,
				authenticatorData: credential.response.authenticatorData
					? toBase64Url(credential.response.authenticatorData)
					: undefined,
				signature: credential.response.signature ? toBase64Url(credential.response.signature) : undefined,
				userHandle: credential.response.userHandle ? toBase64Url(credential.response.userHandle) : undefined,
			},
			clientExtensionResults: credential.getClientExtensionResults(),
		};
	}

	async function postJSON(url, payload, csrf) {
		const response = await fetch(url, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				"X-CSRF-Token": csrf || "",
			},
			body: JSON.stringify(payload),
		});
		if (!response.ok) {
			throw new Error(await response.text());
		}
		return response.json();
	}

	async function handleSignup(form) {
		const displayName = form.querySelector("[name='displayName']").value;
		const restaurantName = form.querySelector("[name='restaurantName']").value;
		const csrf = form.dataset.csrf || "";

		const begin = await postJSON("/auth/register/begin", { displayName, restaurantName }, csrf);
		const creationOptions = begin.publicKey || begin.response || begin;
		const credential = await navigator.credentials.create({
			publicKey: normalizeCreationOptions(creationOptions),
		});
		const finish = await postJSON("/auth/register/finish", serializeCredential(credential), csrf);
		window.location.href = finish.redirect || "/dashboard";
	}

	async function handleInvite(form) {
		const displayName = form.querySelector("[name='displayName']").value;
		const invitationToken = form.dataset.invitationToken;
		const csrf = form.dataset.csrf || "";

		const begin = await postJSON("/auth/register/begin", { displayName, invitationToken }, csrf);
		const creationOptions = begin.publicKey || begin.response || begin;
		const credential = await navigator.credentials.create({
			publicKey: normalizeCreationOptions(creationOptions),
		});
		const finish = await postJSON("/auth/register/finish", serializeCredential(credential), csrf);
		window.location.href = finish.redirect || "/kitchen";
	}

	async function handleLogin(form) {
		const csrf = form.dataset.csrf || "";
		const begin = await postJSON("/auth/login/begin", {}, csrf);
		const requestOptions = begin.publicKey || begin.response || begin;
		const credential = await navigator.credentials.get({
			publicKey: normalizeRequestOptions(requestOptions),
		});
		const finish = await postJSON("/auth/login/finish", serializeCredential(credential), csrf);
		window.location.href = finish.redirect || "/dashboard";
	}

	function bind(formId, handler) {
		const form = document.getElementById(formId);
		if (!form) {
			return;
		}
		form.addEventListener("submit", async (event) => {
			event.preventDefault();
			try {
				await handler(form);
			} catch (error) {
				// Keep UX simple for now and rely on browser alert for errors.
				alert(error.message || "Authentication failed");
			}
		});
	}

	bind("passkey-signup-form", handleSignup);
	bind("passkey-login-form", handleLogin);
	bind("passkey-invite-form", handleInvite);
})();
