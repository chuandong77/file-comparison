export namespace main {
	
	export class requestData {
	    checkType: string;
	    pathA: string;
	    isAppendTimeA: boolean;
	    pathB: string;
	    isAppendTimeB: boolean;
	
	    static createFrom(source: any = {}) {
	        return new requestData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.checkType = source["checkType"];
	        this.pathA = source["pathA"];
	        this.isAppendTimeA = source["isAppendTimeA"];
	        this.pathB = source["pathB"];
	        this.isAppendTimeB = source["isAppendTimeB"];
	    }
	}

}

