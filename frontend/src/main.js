import './style.css';
import './app.css';

import {EventsOn} from '../wailsjs/runtime/runtime';
import {renderDashboard} from './views/dashboard.js';
import {renderDetail} from './views/detail.js';
import {renderSettings} from './views/settings.js';

let current = null;

function showDashboard() {
    current?.destroy?.();
    current = renderDashboard({onOpenTool: showDetail, onOpenSettings: showSettings});
}

function showDetail(toolId) {
    current?.destroy?.();
    current = renderDetail(toolId, {onBack: showDashboard});
}

function showSettings() {
    current?.destroy?.();
    current = renderSettings({onBack: showDashboard});
}

if (window.runtime?.EventsOn) EventsOn('reminder', () => current?.onReminder?.());

showDashboard();
