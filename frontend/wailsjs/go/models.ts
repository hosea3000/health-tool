export namespace domain {
	
	export class Rule {
	    type: string;
	    target?: string;
	    day?: number;
	    weekday?: number;
	    phase?: string;
	    anchor?: string;
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.target = source["target"];
	        this.day = source["day"];
	        this.weekday = source["weekday"];
	        this.phase = source["phase"];
	        this.anchor = source["anchor"];
	    }
	}

}

export namespace model {
	
	export class AppStatus {
	    state: string;
	    elapsedSeconds: number;
	    reminderMinutes: number;
	    restMinutes: number;
	    restRemainingSeconds: number;
	    notificationsEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.elapsedSeconds = source["elapsedSeconds"];
	        this.reminderMinutes = source["reminderMinutes"];
	        this.restMinutes = source["restMinutes"];
	        this.restRemainingSeconds = source["restRemainingSeconds"];
	        this.notificationsEnabled = source["notificationsEnabled"];
	    }
	}
	export class CountdownView {
	    id: string;
	    title: string;
	    rule: domain.Rule;
	    nextDate: string;
	    remainingDays: number;
	
	    static createFrom(source: any = {}) {
	        return new CountdownView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.rule = this.convertValues(source["rule"], domain.Rule);
	        this.nextDate = source["nextDate"];
	        this.remainingDays = source["remainingDays"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CounterHistoryItem {
	    label: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new CounterHistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.count = source["count"];
	    }
	}
	export class CounterView {
	    id: string;
	    name: string;
	    period: string;
	    periodLabel: string;
	    goal: number;
	    count: number;
	    goalReached: boolean;
	    history: CounterHistoryItem[];
	
	    static createFrom(source: any = {}) {
	        return new CounterView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.period = source["period"];
	        this.periodLabel = source["periodLabel"];
	        this.goal = source["goal"];
	        this.count = source["count"];
	        this.goalReached = source["goalReached"];
	        this.history = this.convertValues(source["history"], CounterHistoryItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PendingUpdateInfo {
	    exists: boolean;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new PendingUpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exists = source["exists"];
	        this.version = source["version"];
	    }
	}
	export class Settings {
	    reminderMinutes: number;
	    restMinutes: number;
	    notificationsEnabled: boolean;
	    autoStart?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reminderMinutes = source["reminderMinutes"];
	        this.restMinutes = source["restMinutes"];
	        this.notificationsEnabled = source["notificationsEnabled"];
	        this.autoStart = source["autoStart"];
	    }
	}
	export class TimelineEntry {
	    kind: string;
	    // Go type: time
	    startedAt: any;
	    // Go type: time
	    endedAt?: any;
	    durationSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new TimelineEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.endedAt = this.convertValues(source["endedAt"], null);
	        this.durationSeconds = source["durationSeconds"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateCheckResult {
	    status: string;
	    currentVersion: string;
	    latestVersion: string;
	    releaseUrl: string;
	    downloadUrl: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseUrl = source["releaseUrl"];
	        this.downloadUrl = source["downloadUrl"];
	        this.message = source["message"];
	    }
	}

}

