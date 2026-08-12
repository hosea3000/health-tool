import {registry} from '../tools/index.js';

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
    const slots = {};
    for (const tool of Object.values(registry)) {
        const slot = document.createElement('div');
        slot.className = 'card-slot';
        slot.dataset.toolSlot = tool.id;
        grid.appendChild(slot);
        slots[tool.id] = slot;
    }

    async function renderToolCards(tool) {
        const slot = slots[tool.id];
        if (!slot || !slot.isConnected) return;
        const cards = await tool.renderCards();
        if (cards.length === 0) {
            const placeholder = document.createElement('article');
            placeholder.className = 'tool-card tool-card-empty';
            placeholder.dataset.tool = tool.id;
            placeholder.setAttribute('role', 'button');
            placeholder.tabIndex = 0;
            placeholder.innerHTML = `
                <div class="tool-card-top"><span class="tool-card-kicker">${tool.name}</span><span class="tool-card-status">暂无内容</span></div>
                <div class="tool-card-value">—</div>
                <div class="tool-card-meta"><span>点击进入</span><span>添加内容</span></div>
            `;
            slot.replaceChildren(placeholder);
            return;
        }
        slot.replaceChildren(...cards);
    }

    for (const tool of Object.values(registry)) {
        renderToolCards(tool);
    }

    grid.addEventListener('click', (event) => {
        const card = event.target.closest('[data-tool]');
        if (card) onOpenTool(card.dataset.tool);
    });
    grid.addEventListener('keydown', (event) => {
        if (event.key === 'Enter' && event.target.dataset?.tool) onOpenTool(event.target.dataset.tool);
    });

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
