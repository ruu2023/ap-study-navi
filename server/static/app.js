// Show server-side validation errors from htmx requests (htmx does not swap 4xx/5xx).
document.addEventListener("htmx:responseError", (event) => {
  const flash = document.getElementById("flash");
  if (!flash) return;
  const xhr = event.detail.xhr;
  flash.textContent = (xhr && xhr.status === 401) ? "ログインが切れました。再度ログインしてください。" : (xhr && xhr.responseText) || "保存に失敗しました。";
  flash.hidden = false;
  window.scrollTo({ top: 0 });
});
document.addEventListener("htmx:sendError", () => {
  const flash = document.getElementById("flash");
  if (!flash) return;
  flash.textContent = "通信に失敗しました。接続を確認してください。";
  flash.hidden = false;
});
