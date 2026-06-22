export namespace backend {
	
	export class PostPreview {
	    url: string;
	    thumbnail: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new PostPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.thumbnail = source["thumbnail"];
	        this.kind = source["kind"];
	    }
	}
	export class Result {
	    media_id: string;
	    kind: string;
	    files: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.media_id = source["media_id"];
	        this.kind = source["kind"];
	        this.files = source["files"];
	        this.message = source["message"];
	    }
	}

}

export namespace main {
	
	export class BatchRequest {
	    urls: string[];
	    output: string;
	    hd: boolean;
	    watermark: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BatchRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.urls = source["urls"];
	        this.output = source["output"];
	        this.hd = source["hd"];
	        this.watermark = source["watermark"];
	    }
	}
	export class DownloadRequest {
	    url: string;
	    output: string;
	    hd: boolean;
	    watermark: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DownloadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.output = source["output"];
	        this.hd = source["hd"];
	        this.watermark = source["watermark"];
	    }
	}
	export class HistoryItem {
	    url: string;
	    media_id: string;
	    kind: string;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.media_id = source["media_id"];
	        this.kind = source["kind"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class ProfileRequest {
	    username: string;
	    output: string;
	    hd: boolean;
	    watermark: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.output = source["output"];
	        this.hd = source["hd"];
	        this.watermark = source["watermark"];
	    }
	}

}

