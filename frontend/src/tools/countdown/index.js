import {AddCountdown as backendAdd, CountdownEvents as backendList, DeleteCountdown as backendDelete, UpdateCountdown as backendUpdate} from '../../../wailsjs/go/main/App';

const WEEKDAYS = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
const hasWailsBridge = typeof window.go?.main?.App?.CountdownEvents === 'function';

const api = {
    list: () => hasWailsBridge ? backendList() : Promise.resolve([]),
    add: async (title, rule) => {
        if (!hasWailsBridge) return {ok: false, message: ''};
        const message = await backendAdd(title, rule);
        return {ok: !message, message};
    },
    update: async (id, title, rule) => {
        if (!hasWailsBridge) return {ok: false, message: ''};
        const message = await backendUpdate(id, title, rule);
        return {ok: !message, message};
    },
    remove: (id) => hasWailsBridge ? backendDelete(id) : Promise.resolve(false),
};

function escapeHtml(value) {
    return value.replace(/[&<>"']/g, (c) => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
}

function formatDate(value) {
    const [, m, d] = value.split('-');
    return `${Number(m)}月${Number(d)}日`;
}

function ruleText(rule) {
    switch (rule.type) {
        case 'date':
            return formatDate(rule.target);
        case 'monthly':
            return `每月${rule.day}号`;
        case 'weekly':
            return `每周${WEEKDAYS[rule.weekday]}`;
        case 'biweekly':
            return `${rule.phase === 'big' ? '大周' : '小周'}${WEEKDAYS[rule.weekday]}`;
        default:
            return '';
    }
}

function display(event) {
    const days = event.remainingDays;
    let daysLabel;
    if (days === 0) daysLabel = '今天';
    else if (days > 0) daysLabel = `还剩 ${days} 天`;
    else daysLabel = `已经 ${Math.abs(days)} 天`;
    const prefix = days < 0 ? '已于' : '下次';
    return {value: Math.abs(days), daysLabel, nextLabel: `${prefix} ${formatDate(event.nextDate)}`};
}

function todayISO() {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function weekOptions(selected) {
    return WEEKDAYS.map((name, i) => `<option value="${i}" ${i === selected ? 'selected' : ''}>${name}</option>`).join('');
}

async function renderCards() {
    const events = await api.list();
    return events.map((event) => {
        const view = display(event);
        const card = document.createElement('article');
        card.className = 'tool-card';
        card.dataset.tool = countdownTool.id;
        card.dataset.card = `${countdownTool.id}:${event.id}`;
        card.setAttribute('role', 'button');
        card.tabIndex = 0;
        card.innerHTML = `
            <div class="tool-card-top"><span class="tool-card-kicker">倒数日</span><span class="tool-card-status">${escapeHtml(event.title)}</span></div>
            <div class="tool-card-value">${view.value}</div>
            <div class="tool-card-meta"><span>${ruleText(event.rule)}</span><span>${view.nextLabel}</span></div>
        `;
        return card;
    });
}

function renderDetail(host) {
    host.innerHTML = `
        <section class="countdown-detail">
            <div class="countdown-heading"><div><p class="eyebrow">倒数日</p><h2>重要日子的倒计时</h2></div><button class="button-primary" id="countdown-add">＋ 新增</button></div>
            <div class="countdown-list" id="countdown-list"></div>
        </section>
        <div class="settings-backdrop" id="countdown-form-panel" hidden>
            <section class="settings-card" role="dialog" aria-modal="true" aria-labelledby="countdown-form-title">
                <div class="settings-header"><div><p class="eyebrow">配置</p><h2 id="countdown-form-title">新增倒数日</h2></div><button class="close-settings" id="countdown-form-close">×</button></div>
                <label class="setting-label" for="countdown-title">标题</label>
                <div class="countdown-title-input"><input id="countdown-title" type="text" maxlength="20" placeholder="例如：发薪日"></div>
                <label class="setting-label" for="countdown-rule-type">到期规则</label>
                <select class="countdown-select" id="countdown-rule-type">
                    <option value="date">具体日期</option>
                    <option value="monthly">每月几号</option>
                    <option value="weekly">每周周几</option>
                    <option value="biweekly">大小周周几</option>
                </select>
                <div id="countdown-fields"></div>
                <p class="settings-error" id="countdown-error" hidden>请填写完整的配置信息。</p>
                <div class="settings-actions"><button class="button-secondary" id="countdown-form-cancel">取消</button><button class="button-primary" id="countdown-form-save">保存</button></div>
            </section>
        </div>
        <div class="settings-backdrop" id="countdown-delete-panel" hidden>
            <section class="settings-card" role="dialog" aria-modal="true" aria-labelledby="countdown-delete-title">
                <div class="settings-header"><div><p class="eyebrow">确认</p><h2 id="countdown-delete-title">删除倒数日</h2></div><button class="close-settings" id="countdown-delete-close">×</button></div>
                <p class="settings-description" id="countdown-delete-text">确定删除这个倒数日吗？</p>
                <div class="settings-actions"><button class="button-secondary" id="countdown-delete-cancel">取消</button><button class="button-primary" id="countdown-delete-confirm">删除</button></div>
            </section>
        </div>
    `;

    const list = host.querySelector('#countdown-list');
    const addButton = host.querySelector('#countdown-add');
    const formPanel = host.querySelector('#countdown-form-panel');
    const formTitle = host.querySelector('#countdown-form-title');
    const titleInput = host.querySelector('#countdown-title');
    const ruleType = host.querySelector('#countdown-rule-type');
    const fieldsBox = host.querySelector('#countdown-fields');
    const error = host.querySelector('#countdown-error');
    const deletePanel = host.querySelector('#countdown-delete-panel');
    const deleteText = host.querySelector('#countdown-delete-text');
    let editingId = null;
    let editingAnchor = null;
    let deleteId = null;
    let toastTimer = null;

    function showToast(message) {
        let toast = host.querySelector('.countdown-toast');
        if (!toast) {
            toast = document.createElement('div');
            toast.className = 'countdown-toast';
            host.appendChild(toast);
        }
        toast.textContent = message;
        toast.classList.add('show');
        clearTimeout(toastTimer);
        toastTimer = setTimeout(() => toast.classList.remove('show'), 1600);
    }

    function fieldsHTML(type) {
        switch (type) {
            case 'date':
                return `<label class="setting-label" for="field-date">日期</label>
                        <div class="countdown-inline"><input class="countdown-select countdown-date-input" id="field-date" type="date"></div>`;
            case 'monthly':
                return `<label class="setting-label" for="field-day">每月</label>
                        <div class="countdown-inline"><input class="countdown-number" id="field-day" type="number" min="1" max="31" value="1"><span>号</span></div>`;
            case 'weekly':
                return `<label class="setting-label" for="field-weekday">每周</label>
                        <select class="countdown-select" id="field-weekday">${weekOptions()}</select>`;
            case 'biweekly':
                return `<label class="setting-label">大小周</label>
                        <div class="countdown-inline">
                            <select class="countdown-select" id="field-phase"><option value="big">大周</option><option value="small">小周</option></select>
                            <select class="countdown-select" id="field-bweekday">${weekOptions()}</select>
                        </div>
                        <p class="settings-hint">以创建日所在周为锚，本周固定为大周；选「小周」则从下周开始，之后每两周一次。</p>`;
            default:
                return '';
        }
    }

    function openForm(event = null) {
        editingId = event?.id ?? null;
        editingAnchor = event?.rule?.anchor ?? null;
        formTitle.textContent = editingId ? '编辑倒数日' : '新增倒数日';
        titleInput.value = event?.title ?? '';
        error.hidden = true;
        const type = event?.rule?.type ?? 'date';
        ruleType.value = type;
        fieldsBox.innerHTML = fieldsHTML(type);
        if (type === 'date' && event) fieldsBox.querySelector('#field-date').value = event.rule.target;
        if (type === 'monthly' && event) fieldsBox.querySelector('#field-day').value = event.rule.day;
        if (type === 'weekly' && event) fieldsBox.querySelector('#field-weekday').value = event.rule.weekday;
        if (type === 'biweekly' && event) {
            fieldsBox.querySelector('#field-phase').value = event.rule.phase;
            fieldsBox.querySelector('#field-bweekday').value = event.rule.weekday;
        }
        if (type === 'date' && !event) fieldsBox.querySelector('#field-date').value = todayISO();
        formPanel.hidden = false;
        titleInput.focus();
    }

    function buildRule() {
        const type = ruleType.value;
        switch (type) {
            case 'date':
                return {type, target: fieldsBox.querySelector('#field-date').value};
            case 'monthly':
                return {type, day: Number(fieldsBox.querySelector('#field-day').value)};
            case 'weekly':
                return {type, weekday: Number(fieldsBox.querySelector('#field-weekday').value)};
            case 'biweekly':
                return {type, weekday: Number(fieldsBox.querySelector('#field-bweekday').value), phase: fieldsBox.querySelector('#field-phase').value, anchor: editingAnchor || todayISO()};
            default:
                return null;
        }
    }

    async function saveForm() {
        const title = titleInput.value.trim();
        const rule = buildRule();
        if (!title || !rule || (rule.type === 'date' && !rule.target)) {
            error.textContent = '请填写完整的配置信息。';
            error.hidden = false;
            return;
        }
        const result = editingId ? await api.update(editingId, title, rule) : await api.add(title, rule);
        if (!result.ok) {
            error.textContent = result.message || '保存失败，请检查配置。';
            error.hidden = false;
            return;
        }
        formPanel.hidden = true;
        await refresh();
        showToast('保存成功');
    }

    function closeForm() { formPanel.hidden = true; }

    function openDelete(id, title) {
        deleteId = id;
        deleteText.textContent = `确定删除「${title}」吗？删除后对应卡片也会消失。`;
        deletePanel.hidden = false;
    }

    function closeDelete() { deletePanel.hidden = true; deleteId = null; }

    async function refresh() {
        const events = await api.list();
        if (!events.length) {
            list.innerHTML = '<p class="timeline-empty">还没有倒数日，点右上角"＋ 新增"配置一个吧。</p>';
            return;
        }
        list.innerHTML = events.map((event) => {
            const view = display(event);
            return `<article class="countdown-entry">
                <div class="countdown-entry-head"><strong>${escapeHtml(event.title)}</strong><span>${ruleText(event.rule)}</span></div>
                <p>${view.daysLabel} · ${view.nextLabel}</p>
                <div class="countdown-entry-actions">
                    <button class="countdown-entry-btn" data-action="edit" data-id="${event.id}" data-title="${escapeHtml(event.title)}">编辑</button>
                    <button class="countdown-entry-btn countdown-entry-btn-danger" data-action="delete" data-id="${event.id}" data-title="${escapeHtml(event.title)}">删除</button>
                </div>
            </article>`;
        }).join('');
    }

    list.addEventListener('click', async (event) => {
        const btn = event.target.closest('[data-action]');
        if (!btn) return;
        if (btn.dataset.action === 'edit') {
            const events = await api.list();
            const target = events.find((e) => e.id === btn.dataset.id);
            if (target) openForm(target);
        } else if (btn.dataset.action === 'delete') {
            openDelete(btn.dataset.id, btn.dataset.title);
        }
    });

    addButton.addEventListener('click', () => openForm());
    host.querySelector('#countdown-form-close').addEventListener('click', closeForm);
    host.querySelector('#countdown-form-cancel').addEventListener('click', closeForm);
    host.querySelector('#countdown-form-save').addEventListener('click', saveForm);
    ruleType.addEventListener('change', () => {
        fieldsBox.innerHTML = fieldsHTML(ruleType.value);
        if (ruleType.value === 'date') fieldsBox.querySelector('#field-date').value = todayISO();
    });
    formPanel.addEventListener('click', (e) => { if (e.target === formPanel) closeForm(); });

    host.querySelector('#countdown-delete-close').addEventListener('click', closeDelete);
    host.querySelector('#countdown-delete-cancel').addEventListener('click', closeDelete);
    host.querySelector('#countdown-delete-confirm').addEventListener('click', async () => {
        if (deleteId && await api.remove(deleteId)) {
            closeDelete();
            await refresh();
            showToast('已删除');
        }
    });
    deletePanel.addEventListener('click', (e) => { if (e.target === deletePanel) closeDelete(); });

    refresh();
    const timer = window.setInterval(refresh, countdownTool.refreshInterval);

    return {
        tick: refresh,
        destroy() { window.clearInterval(timer); },
    };
}

export const countdownTool = {
    id: 'countdown',
    name: '倒数日',
    refreshInterval: 60000,
    renderCards,
    renderDetail,
};
