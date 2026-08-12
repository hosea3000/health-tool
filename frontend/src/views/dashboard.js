import {registry} from '../tools/index.js';
import {GetCardOrder as backendGetOrder, SaveCardOrder as backendSaveOrder} from '../../wailsjs/go/main/App';

const hasWailsBridge = typeof window.go?.main?.App?.GetCardOrder === 'function';

const api = {
    getOrder: () => hasWailsBridge ? backendGetOrder() : Promise.resolve([]),
    saveOrder: (order) => {
        if (!hasWailsBridge) return Promise.resolve(true);
        return backendSaveOrder(order);
    },
};

function makePlaceholder(tool) {
    const placeholder = document.createElement('article');
    placeholder.className = 'tool-card tool-card-empty';
    placeholder.dataset.tool = tool.id;
    placeholder.dataset.card = tool.id;
    placeholder.setAttribute('role', 'button');
    placeholder.tabIndex = 0;
    placeholder.innerHTML = `
        <div class="tool-card-top"><span class="tool-card-kicker">${tool.name}</span><span class="tool-card-status">暂无内容</span></div>
        <div class="tool-card-value">—</div>
        <div class="tool-card-meta"><span>点击进入</span><span>添加内容</span></div>
    `;
    return placeholder;
}

function escapeHtml(value) {
    return String(value).replace(/[&<>"']/g, (c) => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
}

function showDashboardError(err) {
    const grid = document.getElementById('card-grid');
    if (grid) grid.innerHTML = `<p class="timeline-empty">页面加载出错：${escapeHtml(err?.message ?? err)}</p>`;
    console.error('[dashboard]', err);
}

export function renderDashboard({onOpenTool}) {
    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="app-shell">
            <header class="topbar">
                <div class="wordmark"><span class="wordmark-mark">✦</span><span>健康工具箱</span></div>
            </header>
            <main class="dashboard">
                <div class="dashboard-heading">
                    <p class="eyebrow">工具箱</p>
                    <h1>打开你的健康角落</h1>
                </div>
                <div class="card-grid" id="card-grid"></div>
            </main>
            <footer class="footer"><span>为身体留一点空间</span><span>健康工具箱 · 本地运行</span></footer>
        </div>
    `;

    const grid = document.getElementById('card-grid');
    const toolCards = new Map();
    let dragging = false;
    let draggingTool = null;
    let dragKey = null;

    function initTool(tool) {
        if (!toolCards.has(tool.id)) toolCards.set(tool.id, new Map());
    }

    async function renderToolCards(tool) {
        if (dragging && draggingTool === tool.id) return;
        initTool(tool);
        let cards = await tool.renderCards();
        if (cards.length === 0) cards = [makePlaceholder(tool)];
        const map = toolCards.get(tool.id);
        const newKeys = new Set();
        for (const card of cards) {
            const key = card.dataset.card;
            newKeys.add(key);
            card.draggable = true;
            const existing = map.get(key);
            if (existing) {
                existing.replaceWith(card);
            } else {
                grid.appendChild(card);
            }
            map.set(key, card);
        }
        for (const [key, node] of map) {
            if (!newKeys.has(key)) {
                node.remove();
                map.delete(key);
            }
        }
    }

    async function mountCards() {
        const byTool = {};
        for (const tool of Object.values(registry)) {
            let cards = await tool.renderCards();
            if (cards.length === 0) cards = [makePlaceholder(tool)];
            byTool[tool.id] = cards;
            toolCards.set(tool.id, new Map());
        }

        const validKeys = [];
        const keyToElement = new Map();
        for (const tool of Object.values(registry)) {
            for (const card of byTool[tool.id]) {
                const key = card.dataset.card;
                validKeys.push(key);
                keyToElement.set(key, card);
                toolCards.get(tool.id).set(key, card);
            }
        }

        const stored = (await api.getOrder()) ?? [];
        const used = new Set();
        const finalOrder = [];
        for (const key of stored) {
            if (!used.has(key) && keyToElement.has(key)) {
                finalOrder.push(key);
                used.add(key);
            }
        }
        for (const key of validKeys) {
            if (!used.has(key)) {
                finalOrder.push(key);
                used.add(key);
            }
        }

        for (const key of finalOrder) {
            const card = keyToElement.get(key);
            card.draggable = true;
            grid.appendChild(card);
        }
    }

    function persistOrder() {
        const order = [...grid.querySelectorAll('[data-card]')].map((card) => card.dataset.card);
        api.saveOrder(order);
    }

    function clearIndicators() {
        grid.querySelectorAll('.drop-before, .drop-after').forEach((card) => card.classList.remove('drop-before', 'drop-after'));
    }

    function clearDraggingState() {
        grid.querySelectorAll('.dragging').forEach((card) => card.classList.remove('dragging'));
        clearIndicators();
        dragging = false;
        draggingTool = null;
        dragKey = null;
    }

    grid.addEventListener('dragstart', (event) => {
        const card = event.target.closest('[data-card]');
        if (!card) return;
        dragKey = card.dataset.card;
        draggingTool = card.dataset.tool;
        dragging = true;
        card.classList.add('dragging');
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('text/plain', dragKey);
    });

    grid.addEventListener('dragover', (event) => {
        event.preventDefault();
        const target = event.target.closest('[data-card]');
        if (!target || target.dataset.card === dragKey) return;
        event.dataTransfer.dropEffect = 'move';
        const rect = target.getBoundingClientRect();
        const after = event.clientY - rect.top > rect.height / 2;
        clearIndicators();
        target.classList.add(after ? 'drop-after' : 'drop-before');
    });

    grid.addEventListener('drop', (event) => {
        event.preventDefault();
        const target = event.target.closest('[data-card]');
        const source = dragKey && [...grid.children].find((card) => card.dataset.card === dragKey);
        if (!target || !source || target === source) {
            clearDraggingState();
            return;
        }
        const rect = target.getBoundingClientRect();
        const after = event.clientY - rect.top > rect.height / 2;
        if (after) target.after(source);
        else target.before(source);
        clearDraggingState();
        persistOrder();
    });

    grid.addEventListener('dragend', clearDraggingState);

    grid.addEventListener('click', (event) => {
        const card = event.target.closest('[data-tool]');
        if (card) onOpenTool(card.dataset.tool);
    });
    grid.addEventListener('keydown', (event) => {
        if (event.key === 'Enter' && event.target.dataset?.tool) onOpenTool(event.target.dataset.tool);
    });

    mountCards().catch(showDashboardError);

    const timers = new Map();
    for (const tool of Object.values(registry)) {
        timers.set(tool.id, window.setInterval(() => renderToolCards(tool), tool.refreshInterval));
    }

    return {
        destroy() {
            timers.forEach((timer) => window.clearInterval(timer));
            timers.clear();
        },
        onReminder() {
            if (registry.reminder) renderToolCards(registry.reminder);
        },
    };
}
