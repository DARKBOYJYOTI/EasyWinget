export namespace models {
	
	export class AppDetails {
	    id: string;
	    name: string;
	    version: string;
	    publisher: string;
	    author: string;
	    moniker: string;
	    description: string;
	    license: string;
	    copyright: string;
	    homepage: string;
	    releaseNotesUrl: string;
	    tags: string[];
	    installerType: string;
	    installerUrl: string;
	    installerSha256: string;
	
	    static createFrom(source: any = {}) {
	        return new AppDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.publisher = source["publisher"];
	        this.author = source["author"];
	        this.moniker = source["moniker"];
	        this.description = source["description"];
	        this.license = source["license"];
	        this.copyright = source["copyright"];
	        this.homepage = source["homepage"];
	        this.releaseNotesUrl = source["releaseNotesUrl"];
	        this.tags = source["tags"];
	        this.installerType = source["installerType"];
	        this.installerUrl = source["installerUrl"];
	        this.installerSha256 = source["installerSha256"];
	    }
	}
	export class AppItem {
	    name: string;
	    id: string;
	    version: string;
	    available?: string;
	    source?: string;
	
	    static createFrom(source: any = {}) {
	        return new AppItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.id = source["id"];
	        this.version = source["version"];
	        this.available = source["available"];
	        this.source = source["source"];
	    }
	}
	export class FileItem {
	    name: string;
	    length: number;
	    lastWriteTime: string;
	
	    static createFrom(source: any = {}) {
	        return new FileItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.length = source["length"];
	        this.lastWriteTime = source["lastWriteTime"];
	    }
	}
	export class IgnoredItem {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new IgnoredItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class JobStatus {
	    id: string;
	    command: string;
	    args: string[];
	    done: boolean;
	    exitCode: number;
	    output: string;
	    // Go type: time
	    startTime: any;
	    completedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new JobStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.done = source["done"];
	        this.exitCode = source["exitCode"];
	        this.output = source["output"];
	        this.startTime = this.convertValues(source["startTime"], null);
	        this.completedAt = source["completedAt"];
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
	export class ManifestIcon {
	    success: boolean;
	    domain: string;
	    url: string;
	    iconUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new ManifestIcon(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.domain = source["domain"];
	        this.url = source["url"];
	        this.iconUrl = source["iconUrl"];
	    }
	}

}

