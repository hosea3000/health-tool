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

}

