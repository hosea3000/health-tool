import {reminderTool} from './reminder/index.js';
import {countdownTool} from './countdown/index.js';
import {counterTool} from './counter/index.js';

export const registry = {
    [reminderTool.id]: reminderTool,
    [countdownTool.id]: countdownTool,
    [counterTool.id]: counterTool,
};
