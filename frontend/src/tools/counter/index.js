import {Counters as backendList, AddCounter as backendAdd, UpdateCounter as backendUpdate, DeleteCounter as backendDelete, IncrementCounter as backendIncrement, DecrementCounter as backendDecrement, SetCounterCount as backendSet} from '../../../wailsjs/go/main/App';

const PERIOD_NAMES = {day: '每天', month: '每月', year: '每年', never: '永不清零'};
const hasWailsBridge = typeof window.go?.main?.App?.Counters === 'function';

const api = {
    list: () => hasWailsBridge ? backendList() : Promise.resolve([]),
    add: async (name, period, goal) => {
        if (!hasWailsBridge) return {ok: false, message: ''};
        const message = await backendAdd(name, period, goal);
        return {ok: !message, message};
    },
    update: async (id, name, period, goal) => {
        if (!hasWailsBridge) return {ok: false, message: ''};
        const message = await backendUpdate(id, name, period, goal);
        return {ok: !message, message};
    },
    remove: (id) => hasWailsBridge ? backendDelete(id) : Promise.resolve(false),
    increment: (id) => hasWailsBridge ? backendIncrement(id) : Promise.resolve(0),
    decrement: (id) => hasWailsBridge ? backendDecrement(id) : Promise.resolve(0),
    setCount: (id, count) => hasWailsBridge ? backendSet(id, count) : Promise.resolve(0),
};

function escapeHtml(value) {
    return String(value).replace(/[&<>"']/g, (c) => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
}

function goalText(counter) {
    if (counter.goal > 0) return counter.count >= counter.goal ? '已达成' : `还差 ${counter.goal - counter.count} 次`;
    return `${counter.count} 次`;
}

function cardHTML(counter) {
    return `
        <div class="tool-card-top">
            <span class="tool-card-kicker">计数器</span>
            <span class="tool-card-head-right">
                <span class="tool-card-status">${escapeHtml(counter.name)}</span>
                <button class="counter-card-add" aria-label="加一次">＋</button>
            </span>
        </div>
        <div class="tool-card-value">${counter.count}</div>
        <div class="tool-card-meta"><span>${counter.periodLabel}</span><span>${goalText(counter)}</span></div>
    `;
}

function buildCard(counter) {
    const card = document.createElement('article');
    card.className = 'tool-card';
    card.dataset.tool = counterTool.id;
    card.dataset.card = `${counterTool.id}:${counter.id}`;
    card.setAttribute('role', 'button');
    card.tabIndex = 0;
    card.innerHTML = cardHTML(counter);
    card.addEventListener('click', async (event) => {
        const addBtn = event.target.closest('.counter-card-add');
        if (!addBtn) return;
        event.stopPropagation();
        const newCount = await api.increment(counter.id);
        if (newCount < 0) return;
        counter.count = newCount;
        card.innerHTML = cardHTML(counter);
    });
    return card;
}

async function renderCards() {
    const counters = await api.list();
    return counters.map(buildCard);
}

function entryHTML(counter) {
    const state = counter.goal > 0 && counter.count >= counter.goal ? '<span>已达成</span>' : '';
    const history = (counter.history || []).map((h) => `<span class="counter-entry-history-item">${escapeHtml(h.label)} · ${h.count} 次</span>`).join('');
    return `<article class="counter-entry">
        <div class="counter-entry-head">
            <strong>${escapeHtml(counter.name)}</strong>
            <span>周期：${PERIOD_NAMES[counter.period] ?? counter.period}${counter.goal > 0 ? ` · 目标 ${counter.goal}` : ''}</span>
        </div>
        <div class="counter-entry-controls">
            <button class="counter-entry-btn counter-entry-step" data-action="decrement" data-id="${counter.id}" aria-label="减一次">−</button>
            <input class="counter-entry-number" type="number" min="0" value="${counter.count}" data-id="${counter.id}" aria-label="当前次数">
            <button class="counter-entry-btn counter-entry-step" data-action="increment" data-id="${counter.id}" aria-label="加一次">＋</button>
        </div>
        <div class="counter-entry-meta"><span>${counter.periodLabel} · ${counter.count} 次</span>${state}</div>
        ${history ? `<div class="counter-entry-history">${history}</div>` : ''}
        <div class="counter-entry-actions">
            <button class="counter-entry-btn" data-action="edit" data-id="${counter.id}" data-name="${escapeHtml(counter.name)}">编辑</button>
            <button class="counter-entry-btn counter-entry-btn-danger" data-action="delete" data-id="${counter.id}" data-name="${escapeHtml(counter.name)}">删除</button>
        </div>
    </article>`;
}

function renderDetail(host) {
    host.innerHTML = `
        <section class="counter-detail">
            <div class="counter-heading"><div><p class="eyebrow">计数器</p><h2>记录今天的小习惯</h2></div><button class="button-primary" id="counter-add">＋ 新增</button></div>
            <div class="counter-list" id="counter-list"></div>
        </section>
        <div class="settings-backdrop" id="counter-form-panel" hidden>
            <section class="settings-card" role="dialog" aria-modal="true" aria-labelledby="counter-form-title">
                <div class="settings-header"><div><p class="eyebrow">配置</p><h2 id="counter-form-title">新增计数器</h2></div><button class="close-settings" id="counter-form-close">×</button></div>
                <label class="setting-label" for="counter-name">名称</label>
                <div class="countdown-title-input"><input id="counter-name" type="text" maxlength="20" placeholder="例如：喝水"></div>
                <label class="setting-label" for="counter-period">重置周期</label>
                <select class="countdown-select" id="counter-period">
                    <option value="day">每天</option>
                    <option value="month">每月</option>
                    <option value="year">每年</option>
                    <option value="never">永不清零</option>
                </select>
                <label class="setting-label" for="counter-goal">目标值（可留空）</label>
                <div class="settings-value"><input id="counter-goal" type="number" min="0" value="0"><span>次</span></div>
                <p class="settings-hint">设置目标后，卡片会显示"还差 N 次 / 已达成"；留空则只计数不设上限。</p>
                <p class="settings-error" id="counter-error" hidden>请填写完整的配置信息。</p>
                <div class="settings-actions"><button class="button-secondary" id="counter-form-cancel">取消</button><button class="button-primary" id="counter-form-save">保存</button></div>
            </section>
        </div>
        <div class="settings-backdrop" id="counter-delete-panel" hidden>
            <section class="settings-card" role="dialog" aria-modal="true" aria-labelledby="counter-delete-title">
                <div class="settings-header"><div><p class="eyebrow">确认</p><h2 id="counter-delete-title">删除计数器</h2></div><button class="close-settings" id="counter-delete-close">×</button></div>
                <p class="settings-description" id="counter-delete-text">确定删除这个计数器吗？</p>
                <div class="settings-actions"><button class="button-secondary" id="counter-delete-cancel">取消</button><button class="button-primary" id="counter-delete-confirm">删除</button></div>
            </section>
        </div>
    `;

    const list = host.querySelector('#counter-list');
    const addButton = host.querySelector('#counter-add');
    const formPanel = host.querySelector('#counter-form-panel');
    const formTitle = host.querySelector('#counter-form-title');
    const nameInput = host.querySelector('#counter-name');
    const periodSelect = host.querySelector('#counter-period');
    const goalInput = host.querySelector('#counter-goal');
    const error = host.querySelector('#counter-error');
    const deletePanel = host.querySelector('#counter-delete-panel');
    const deleteText = host.querySelector('#counter-delete-text');
    let editingId = null;
    let deleteId = null;
    let toastTimer = null;

    function showToast(message) {
        let toast = host.querySelector('.counter-toast');
        if (!toast) {
            toast = document.createElement('div');
            toast.className = 'counter-toast';
            host.appendChild(toast);
        }
        toast.textContent = message;
        toast.classList.add('show');
        clearTimeout(toastTimer);
        toastTimer = setTimeout(() => toast.classList.remove('show'), 1600);
    }

    function openForm(counter = null) {
        editingId = counter?.id ?? null;
        formTitle.textContent = editingId ? '编辑计数器' : '新增计数器';
        nameInput.value = counter?.name ?? '';
        periodSelect.value = counter?.period ?? 'day';
        goalInput.value = counter?.goal ?? 0;
        error.hidden = true;
        formPanel.hidden = false;
        nameInput.focus();
    }

    function closeForm() { formPanel.hidden = true; }

    function buildPayload() {
        const name = nameInput.value.trim();
        const period = periodSelect.value;
        const goal = Math.max(0, Math.floor(Number(goalInput.value) || 0));
        return {name, period, goal};
    }

    async function saveForm() {
        const {name, period, goal} = buildPayload();
        if (!name) {
            error.textContent = '请填写名称。';
            error.hidden = false;
            return;
        }
        const result = editingId ? await api.update(editingId, name, period, goal) : await api.add(name, period, goal);
        if (!result.ok) {
            error.textContent = result.message || '保存失败，请检查配置。';
            error.hidden = false;
            return;
        }
        formPanel.hidden = true;
        await refresh();
        showToast('保存成功');
    }

    function openDelete(id, name) {
        deleteId = id;
        deleteText.textContent = `确定删除「${name}」吗？删除后对应卡片也会消失。`;
        deletePanel.hidden = false;
    }

    function closeDelete() { deletePanel.hidden = true; deleteId = null; }

    async function refresh() {
        const counters = await api.list();
        if (!counters.length) {
            list.innerHTML = '<p class="timeline-empty">还没有计数器，点右上角"＋ 新增"配置一个吧。</p>';
            return;
        }
        list.innerHTML = counters.map(entryHTML).join('');
    }

    list.addEventListener('click', async (event) => {
        const btn = event.target.closest('[data-action]');
        if (!btn) return;
        const {action, id, name} = btn.dataset;
        if (action === 'increment') {
            await api.increment(id);
            await refresh();
        } else if (action === 'decrement') {
            await api.decrement(id);
            await refresh();
        } else if (action === 'edit') {
            const counters = await api.list();
            const target = counters.find((c) => c.id === id);
            if (target) openForm(target);
        } else if (action === 'delete') {
            openDelete(id, name);
        }
    });

    list.addEventListener('change', async (event) => {
        const input = event.target.closest('.counter-entry-number');
        if (!input) return;
        const count = Math.max(0, Math.floor(Number(input.value) || 0));
        await api.setCount(input.dataset.id, count);
        await refresh();
    });

    addButton.addEventListener('click', () => openForm());
    host.querySelector('#counter-form-close').addEventListener('click', closeForm);
    host.querySelector('#counter-form-cancel').addEventListener('click', closeForm);
    host.querySelector('#counter-form-save').addEventListener('click', saveForm);
    formPanel.addEventListener('click', (e) => { if (e.target === formPanel) closeForm(); });

    host.querySelector('#counter-delete-close').addEventListener('click', closeDelete);
    host.querySelector('#counter-delete-cancel').addEventListener('click', closeDelete);
    host.querySelector('#counter-delete-confirm').addEventListener('click', async () => {
        if (deleteId && await api.remove(deleteId)) {
            closeDelete();
            await refresh();
            showToast('已删除');
        }
    });
    deletePanel.addEventListener('click', (e) => { if (e.target === deletePanel) closeDelete(); });

    refresh();
    const timer = window.setInterval(async () => {
        if (document.activeElement?.classList?.contains('counter-entry-number')) return;
        await refresh();
    }, counterTool.refreshInterval);

    return {
        tick: refresh,
        destroy() { window.clearInterval(timer); },
    };
}

export const counterTool = {
    id: 'counter',
    name: '计数器',
    refreshInterval: 60000,
    renderCards,
    renderDetail,
};
