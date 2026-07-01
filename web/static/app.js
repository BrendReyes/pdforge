'use strict';

// ── Tab switching ─────────────────────────────────────────────────────────────

document.querySelectorAll('.tab').forEach(function(tab) {
  tab.addEventListener('click', function() {
    document.querySelectorAll('.tab').forEach(function(t) { t.classList.remove('active'); });
    document.querySelectorAll('.op-panel').forEach(function(p) { p.classList.remove('active'); });
    tab.classList.add('active');
    document.getElementById('op-' + tab.dataset.op).classList.add('active');
    resetStatus();
  });
});

// ── Form wiring ───────────────────────────────────────────────────────────────

['optimize', 'merge', 'convert', 'rotate', 'rmpage', 'split'].forEach(function(op) {
  var form = document.getElementById('form-' + op);
  if (form) {
    form.addEventListener('submit', function(e) { handleSubmit(e, op); });
  }
});

function handleSubmit(e, op) {
  e.preventDefault();
  var form = e.target;
  var data = new FormData(form);

  // Checkboxes in split are submitted as 'on' or absent by FormData;
  // the server expects the string "true" or "false".
  if (op === 'split') {
    data.set('extract', form.querySelector('[name=extract]').checked ? 'true' : 'false');
    data.set('odd',     form.querySelector('[name=odd]').checked     ? 'true' : 'false');
    data.set('even',    form.querySelector('[name=even]').checked    ? 'true' : 'false');
  }

  var btn = form.querySelector('.btn-submit');
  btn.disabled = true;
  showProgress('Uploading…');

  fetch('/api/' + op, { method: 'POST', body: data })
    .then(function(res) { return res.json(); })
    .then(function(body) {
      if (body.error) throw new Error(body.error);
      startSSE(body.token, btn);
    })
    .catch(function(err) {
      showError(err.message || 'Upload failed');
      btn.disabled = false;
    });
}

// ── SSE progress ──────────────────────────────────────────────────────────────

function startSSE(token, btn) {
  showProgress('Processing…');

  var es = new EventSource('/api/progress/' + token);

  es.onmessage = function(e) {
    var event;
    try { event = JSON.parse(e.data); } catch (_) { return; }

    switch (event.type) {
      case 'started':
        showProgress('Processing…');
        break;
      case 'done':
        es.close();
        showDone(token);
        btn.disabled = false;
        break;
      case 'error':
        es.close();
        showError(event.message || 'Operation failed');
        btn.disabled = false;
        break;
    }
  };

  es.onerror = function() {
    es.close();
    showError('Connection lost — the server may have stopped');
    btn.disabled = false;
  };
}

// ── Status helpers ────────────────────────────────────────────────────────────

var statusArea   = document.getElementById('status-area');
var statusText   = document.getElementById('status-text');
var progressFill = document.getElementById('progress-fill');
var resultArea   = document.getElementById('result-area');
var downloadLink = document.getElementById('download-link');
var errorArea    = document.getElementById('error-area');
var errorText    = document.getElementById('error-text');

function resetStatus() {
  statusArea.classList.add('hidden');
  resultArea.classList.add('hidden');
  errorArea.classList.add('hidden');
  progressFill.classList.remove('done');
}

function showProgress(msg) {
  statusArea.classList.remove('hidden');
  resultArea.classList.add('hidden');
  errorArea.classList.add('hidden');
  progressFill.classList.remove('done');
  statusText.textContent = msg;
}

function showDone(token) {
  progressFill.classList.add('done');
  statusText.textContent = 'Done';
  downloadLink.href = '/api/download/' + token;
  resultArea.classList.remove('hidden');
  errorArea.classList.add('hidden');
}

function showError(msg) {
  progressFill.classList.remove('done');
  statusText.textContent = 'Failed';
  errorText.textContent = msg;
  errorArea.classList.remove('hidden');
  resultArea.classList.add('hidden');
}
