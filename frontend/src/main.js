import './style.css';
import './app.css';

import {EventsOn} from '../wailsjs/runtime/runtime';
import {renderDashboard} from './views/dashboard.js';
import {renderDetail} from './views/detail.js';

let current = null;

function showDashboard() {
    current?.destroy?.();
    current = renderDashboard({onOpenTool: showDetail});
}

function showDetail(toolId) {
    current?.destroy?.();
    current = renderDetail(toolId, {onBack: showDashboard});
}

if (window.runtime?.EventsOn) EventsOn('reminder', () => current?.onReminder?.());

showDashboard();
