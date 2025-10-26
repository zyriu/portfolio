export namespace backend {
	
	export class JobLog {
	    timestamp: string;
	    message: string;
	    level: string;
	
	    static createFrom(source: any = {}) {
	        return new JobLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.message = source["message"];
	        this.level = source["level"];
	    }
	}
	export class JobExecution {
	    id: string;
	    jobName: string;
	    startTime: string;
	    endTime?: string;
	    status: string;
	    logs: JobLog[];
	
	    static createFrom(source: any = {}) {
	        return new JobExecution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.jobName = source["jobName"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.status = source["status"];
	        this.logs = this.convertValues(source["logs"], JobLog);
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
	
	export class JobState {
	    name: string;
	    interval: number;
	    running: boolean;
	    lastRunUnix: number;
	    nextRunUnix: number;
	    err: string;
	    isExecuting: boolean;
	    currentStatus: string;
	    logs: JobLog[];
	
	    static createFrom(source: any = {}) {
	        return new JobState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.interval = source["interval"];
	        this.running = source["running"];
	        this.lastRunUnix = source["lastRunUnix"];
	        this.nextRunUnix = source["nextRunUnix"];
	        this.err = source["err"];
	        this.isExecuting = source["isExecuting"];
	        this.currentStatus = source["currentStatus"];
	        this.logs = this.convertValues(source["logs"], JobLog);
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

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

