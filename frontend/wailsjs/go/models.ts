export namespace backup {
	
	export class Backup {
	    id: string;
	    service: string;
	    // Go type: time
	    timestamp: any;
	    path: string;
	    originalPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new Backup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.service = source["service"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.path = source["path"];
	        this.originalPath = source["originalPath"];
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

export namespace launchctl {
	
	export class CalendarEntry {
	    minute?: number;
	    hour?: number;
	    day?: number;
	    weekday?: number;
	    month?: number;
	
	    static createFrom(source: any = {}) {
	        return new CalendarEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minute = source["minute"];
	        this.hour = source["hour"];
	        this.day = source["day"];
	        this.weekday = source["weekday"];
	        this.month = source["month"];
	    }
	}
	export class ScheduleConfig {
	    schedules?: CalendarEntry[];
	    interval?: number;
	
	    static createFrom(source: any = {}) {
	        return new ScheduleConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schedules = this.convertValues(source["schedules"], CalendarEntry);
	        this.interval = source["interval"];
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
	export class Service {
	    name: string;
	    label: string;
	    status: string;
	    pid?: number;
	    path: string;
	    program?: string;
	    arguments?: string[];
	    runAtLoad: boolean;
	    keepAlive: boolean;
	    schedule?: ScheduleConfig;
	    environment?: Record<string, string>;
	    stdoutPath?: string;
	    stderrPath?: string;
	    wakeSystem: boolean;
	    workingDirectory?: string;
	    type: string;
	    readOnly: boolean;
	    plistFormat: string;
	    statusConfidence: string;
	
	    static createFrom(source: any = {}) {
	        return new Service(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.label = source["label"];
	        this.status = source["status"];
	        this.pid = source["pid"];
	        this.path = source["path"];
	        this.program = source["program"];
	        this.arguments = source["arguments"];
	        this.runAtLoad = source["runAtLoad"];
	        this.keepAlive = source["keepAlive"];
	        this.schedule = this.convertValues(source["schedule"], ScheduleConfig);
	        this.environment = source["environment"];
	        this.stdoutPath = source["stdoutPath"];
	        this.stderrPath = source["stderrPath"];
	        this.wakeSystem = source["wakeSystem"];
	        this.workingDirectory = source["workingDirectory"];
	        this.type = source["type"];
	        this.readOnly = source["readOnly"];
	        this.plistFormat = source["plistFormat"];
	        this.statusConfidence = source["statusConfidence"];
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
	export class ServiceConfig {
	    label: string;
	    program?: string;
	    arguments?: string[];
	    runAtLoad: boolean;
	    keepAlive: boolean;
	    schedule?: ScheduleConfig;
	    environment?: Record<string, string>;
	    wakeSystem: boolean;
	    stdoutPath?: string;
	    stderrPath?: string;
	    workingDirectory?: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.program = source["program"];
	        this.arguments = source["arguments"];
	        this.runAtLoad = source["runAtLoad"];
	        this.keepAlive = source["keepAlive"];
	        this.schedule = this.convertValues(source["schedule"], ScheduleConfig);
	        this.environment = source["environment"];
	        this.wakeSystem = source["wakeSystem"];
	        this.stdoutPath = source["stdoutPath"];
	        this.stderrPath = source["stderrPath"];
	        this.workingDirectory = source["workingDirectory"];
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

export namespace main {
	
	export class AdminModeStatus {
	    state: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AdminModeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.error = source["error"];
	    }
	}

}

export namespace plistutil {
	
	export class Content {
	    data: string;
	    format: string;
	    convertFailed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Content(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.format = source["format"];
	        this.convertFailed = source["convertFailed"];
	    }
	}

}

