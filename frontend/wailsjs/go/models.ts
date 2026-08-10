export namespace main {
	
	export class AppStatus {
	    state: string;
	    elapsedSeconds: number;
	    reminderMinutes: number;
	    restMinutes: number;
	    restRemainingSeconds: number;
	
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
	    }
	}
	export class Settings {
	    reminderMinutes: number;
	    restMinutes: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reminderMinutes = source["reminderMinutes"];
	        this.restMinutes = source["restMinutes"];
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

}

