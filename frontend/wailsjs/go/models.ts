export namespace eve {
	
	export class ESIAuth {
	    access_token: string;
	    refresh_token: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ESIAuth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.access_token = source["access_token"];
	        this.refresh_token = source["refresh_token"];
	        this.name = source["name"];
	    }
	}
	export class FrontendKillmailAttackers {
	    character_id: string;
	    alliance_id: string;
	    corporation_id: string;
	    ship_type_id: string;
	
	    static createFrom(source: any = {}) {
	        return new FrontendKillmailAttackers(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.character_id = source["character_id"];
	        this.alliance_id = source["alliance_id"];
	        this.corporation_id = source["corporation_id"];
	        this.ship_type_id = source["ship_type_id"];
	    }
	}
	export class FrontendKillmailVictim {
	    character_id: string;
	    alliance_id: string;
	    corporation_id: string;
	    ship_type_id: string;
	
	    static createFrom(source: any = {}) {
	        return new FrontendKillmailVictim(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.character_id = source["character_id"];
	        this.alliance_id = source["alliance_id"];
	        this.corporation_id = source["corporation_id"];
	        this.ship_type_id = source["ship_type_id"];
	    }
	}
	export class FrontendKillmail {
	    victim: FrontendKillmailVictim;
	    attackers: FrontendKillmailAttackers[];
	    killmailId: number;
	    killmail_time: string;
	
	    static createFrom(source: any = {}) {
	        return new FrontendKillmail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.victim = this.convertValues(source["victim"], FrontendKillmailVictim);
	        this.attackers = this.convertValues(source["attackers"], FrontendKillmailAttackers);
	        this.killmailId = source["killmailId"];
	        this.killmail_time = source["killmail_time"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	
	
	export class LocationResponse {
	    name: string;
	    solar_system_id: number;
	    station_id: number;
	    structure_id: number;
	
	    static createFrom(source: any = {}) {
	        return new LocationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.solar_system_id = source["solar_system_id"];
	        this.station_id = source["station_id"];
	        this.structure_id = source["structure_id"];
	    }
	}

}

