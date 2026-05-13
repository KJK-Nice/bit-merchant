// Admin item-editor controller: keeps the hidden option_groups_json input in
// sync with the rendered modifier-group cards. Plain DOM — no Datastar.
(function () {
  'use strict';

  var hidden = document.getElementById('ie-option-groups-json');
  var groupsContainer = document.getElementById('ie-mod-groups');
  var addGroupBtn = document.getElementById('ie-add-group');
  if (!hidden || !groupsContainer || !addGroupBtn) return;

  var state = [];
  try {
    var parsed = JSON.parse(hidden.value || '[]');
    if (Array.isArray(parsed)) state = parsed;
  } catch (_) {
    state = [];
  }

  function uid(prefix) {
    return prefix + '_' + Math.random().toString(36).slice(2, 9);
  }

  function sync() {
    // Drop client-only fields and normalise empty optional ones.
    var clean = state.map(function (g) {
      return {
        id: g.id || uid('g'),
        name: g.name || '',
        required: !!g.required,
        min_selections: g.required ? Math.max(1, g.min_selections | 0) : Math.max(0, g.min_selections | 0),
        max_selections: Math.max(0, g.max_selections | 0),
        default_option_id: g.default_option_id || null,
        options: (g.options || []).map(function (o) {
          return {
            id: o.id || uid('o'),
            name: o.name || '',
            price_delta: parseFloat(o.price_delta || 0) || 0,
          };
        }),
      };
    });
    hidden.value = JSON.stringify(clean);
    render();
  }

  function render() {
    groupsContainer.innerHTML = '';
    state.forEach(function (g, gi) {
      groupsContainer.appendChild(renderGroup(g, gi));
    });
  }

  function renderGroup(g, gi) {
    var card = document.createElement('div');
    card.className = 'rounded-lg border border-border p-3 space-y-2';
    card.dataset.groupIndex = String(gi);

    var head = document.createElement('div');
    head.className = 'flex items-center gap-2';

    var nameInput = document.createElement('input');
    nameInput.type = 'text';
    nameInput.placeholder = 'Group name (e.g. Choose a sauce)';
    nameInput.value = g.name || '';
    nameInput.className = 'flex h-9 flex-1 rounded-md border border-input bg-background px-3 py-1 text-sm';
    nameInput.addEventListener('input', function () {
      state[gi].name = nameInput.value;
      hiddenOnly();
    });
    head.appendChild(nameInput);

    var removeBtn = btn('Remove', 'text-destructive text-xs');
    removeBtn.addEventListener('click', function () {
      state.splice(gi, 1);
      sync();
    });
    head.appendChild(removeBtn);
    card.appendChild(head);

    var rules = document.createElement('div');
    rules.className = 'flex flex-wrap items-center gap-3 text-xs';

    var requiredLbl = document.createElement('label');
    requiredLbl.className = 'flex items-center gap-1';
    var requiredInput = document.createElement('input');
    requiredInput.type = 'checkbox';
    requiredInput.checked = !!g.required;
    requiredInput.addEventListener('change', function () {
      state[gi].required = requiredInput.checked;
      if (requiredInput.checked && (state[gi].min_selections | 0) < 1) {
        state[gi].min_selections = 1;
      }
      sync();
    });
    requiredLbl.appendChild(requiredInput);
    requiredLbl.appendChild(document.createTextNode(' Required'));
    rules.appendChild(requiredLbl);

    rules.appendChild(numField('Min', g.min_selections || 0, function (v) {
      state[gi].min_selections = v;
      hiddenOnly();
    }));
    rules.appendChild(numField('Max', g.max_selections || 0, function (v) {
      state[gi].max_selections = v;
      hiddenOnly();
    }));
    card.appendChild(rules);

    var optsContainer = document.createElement('div');
    optsContainer.className = 'space-y-2';
    (g.options || []).forEach(function (o, oi) {
      optsContainer.appendChild(renderOption(gi, oi, o, g.default_option_id));
    });
    card.appendChild(optsContainer);

    var addOpt = btn('+ Add option', 'text-sm');
    addOpt.addEventListener('click', function () {
      state[gi].options = state[gi].options || [];
      state[gi].options.push({ id: uid('o'), name: '', price_delta: 0 });
      sync();
    });
    card.appendChild(addOpt);

    return card;
  }

  function renderOption(gi, oi, o, defaultID) {
    var row = document.createElement('div');
    row.className = 'flex items-center gap-2';

    var nameInput = document.createElement('input');
    nameInput.type = 'text';
    nameInput.value = o.name || '';
    nameInput.placeholder = 'Option name';
    nameInput.className = 'flex h-9 flex-1 rounded-md border border-input bg-background px-3 py-1 text-sm';
    nameInput.addEventListener('input', function () {
      state[gi].options[oi].name = nameInput.value;
      hiddenOnly();
    });
    row.appendChild(nameInput);

    var price = document.createElement('input');
    price.type = 'number';
    price.step = '0.01';
    price.value = String(o.price_delta || 0);
    price.placeholder = '0.00';
    price.className = 'h-9 w-24 rounded-md border border-input bg-background px-3 py-1 text-sm';
    price.addEventListener('input', function () {
      state[gi].options[oi].price_delta = parseFloat(price.value) || 0;
      hiddenOnly();
    });
    row.appendChild(price);

    var defLbl = document.createElement('label');
    defLbl.className = 'flex items-center gap-1 text-xs';
    var defInput = document.createElement('input');
    defInput.type = 'radio';
    defInput.name = 'default_g' + gi;
    defInput.checked = (o.id && o.id === defaultID);
    defInput.addEventListener('change', function () {
      state[gi].default_option_id = o.id;
      hiddenOnly();
    });
    defLbl.appendChild(defInput);
    defLbl.appendChild(document.createTextNode(' default'));
    row.appendChild(defLbl);

    var rm = btn('×', 'text-destructive');
    rm.addEventListener('click', function () {
      state[gi].options.splice(oi, 1);
      if (state[gi].default_option_id === o.id) {
        state[gi].default_option_id = null;
      }
      sync();
    });
    row.appendChild(rm);

    return row;
  }

  function btn(label, extraClass) {
    var b = document.createElement('button');
    b.type = 'button';
    b.textContent = label;
    b.className = 'inline-flex h-8 items-center justify-center rounded-md px-2 ' + (extraClass || '');
    return b;
  }

  function numField(labelText, value, onChange) {
    var wrap = document.createElement('label');
    wrap.className = 'flex items-center gap-1';
    var span = document.createElement('span');
    span.textContent = labelText;
    wrap.appendChild(span);
    var input = document.createElement('input');
    input.type = 'number';
    input.min = '0';
    input.value = String(value);
    input.className = 'h-8 w-14 rounded-md border border-input bg-background px-2 py-1 text-xs';
    input.addEventListener('input', function () {
      onChange(parseInt(input.value, 10) || 0);
    });
    wrap.appendChild(input);
    return wrap;
  }

  // hiddenOnly updates the JSON without re-rendering — used for input events
  // where re-rendering would lose focus.
  function hiddenOnly() {
    hidden.value = JSON.stringify(state.map(function (g) {
      return {
        id: g.id || uid('g'),
        name: g.name || '',
        required: !!g.required,
        min_selections: g.required ? Math.max(1, g.min_selections | 0) : Math.max(0, g.min_selections | 0),
        max_selections: Math.max(0, g.max_selections | 0),
        default_option_id: g.default_option_id || null,
        options: (g.options || []).map(function (o) {
          return {
            id: o.id || uid('o'),
            name: o.name || '',
            price_delta: parseFloat(o.price_delta || 0) || 0,
          };
        }),
      };
    }));
  }

  addGroupBtn.addEventListener('click', function () {
    state.push({ id: uid('g'), name: '', required: false, min_selections: 0, max_selections: 0, default_option_id: null, options: [] });
    sync();
  });

  // Initial render.
  sync();
})();
