// Handles email+password login, signup, and invitation-accept flows.
(function () {
  function csrfToken(form) {
    return form.dataset.csrf || "";
  }

  function collectFields(form, fields) {
    const data = {};
    for (const f of fields) {
      // Prefer field inside this form; fall back to first match in document for shared fields.
      const el = form.querySelector(`[name="${f}"]`) || document.querySelector(`[name="${f}"]`);
      if (el) data[f] = el.value;
    }
    return data;
  }

  async function postJSON(url, body, csrfTok) {
    const resp = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfTok,
      },
      body: JSON.stringify(body),
    });
    if (!resp.ok) {
      const msg = await resp.text();
      throw new Error(msg || resp.statusText);
    }
    return resp.json();
  }

  function showError(form, msg) {
    let el = form.querySelector(".pw-error");
    if (!el) {
      el = document.createElement("p");
      el.className = "pw-error text-sm text-destructive mt-2";
      form.appendChild(el);
    }
    el.textContent = msg;
  }

  // Password login
  const loginForm = document.getElementById("password-login-form");
  if (loginForm) {
    loginForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const data = collectFields(loginForm, ["email", "password"]);
      try {
        const res = await postJSON("/auth/login/password", data, csrfToken(loginForm));
        window.location.href = res.redirect || "/dashboard";
      } catch (err) {
        showError(loginForm, err.message);
      }
    });
  }

  // Password signup (owner)
  const signupForm = document.getElementById("password-signup-form");
  if (signupForm) {
    signupForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const data = collectFields(signupForm, ["displayName", "restaurantName", "email", "password"]);
      try {
        const res = await postJSON("/auth/register/password", data, csrfToken(signupForm));
        window.location.href = res.redirect || "/dashboard";
      } catch (err) {
        showError(signupForm, err.message);
      }
    });
  }

  // Password invitation accept (kitchen staff)
  const inviteForm = document.getElementById("password-invite-form");
  if (inviteForm) {
    inviteForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const data = collectFields(inviteForm, ["displayName", "email", "password"]);
      data.invitationToken = inviteForm.dataset.invitationToken || "";
      try {
        const res = await postJSON("/auth/register/password", data, csrfToken(inviteForm));
        window.location.href = res.redirect || "/dashboard";
      } catch (err) {
        showError(inviteForm, err.message);
      }
    });
  }
})();
