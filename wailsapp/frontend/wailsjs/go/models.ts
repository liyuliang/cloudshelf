export namespace main {
	
	export class AccountDTO {
	    ID: number;
	    Name: string;
	    Provider: string;
	    Endpoint: string;
	    Region: string;
	    AccessKeyID: string;
	    SecretAccessKey: string;
	    Bucket: string;
	    UsePathStyle: boolean;
	    TenantID: string;
	    ClientID: string;
	    ClientSecret: string;
	    SiteURL: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Provider = source["Provider"];
	        this.Endpoint = source["Endpoint"];
	        this.Region = source["Region"];
	        this.AccessKeyID = source["AccessKeyID"];
	        this.SecretAccessKey = source["SecretAccessKey"];
	        this.Bucket = source["Bucket"];
	        this.UsePathStyle = source["UsePathStyle"];
	        this.TenantID = source["TenantID"];
	        this.ClientID = source["ClientID"];
	        this.ClientSecret = source["ClientSecret"];
	        this.SiteURL = source["SiteURL"];
	    }
	}
	export class BucketDTO {
	    Name: string;
	    Accessible: boolean;
	    Checked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BucketDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Accessible = source["Accessible"];
	        this.Checked = source["Checked"];
	    }
	}
	export class DriveDTO {
	    ID: string;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new DriveDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	    }
	}
	export class LANShareDTO {
	    token: string;
	    name: string;
	    provider: string;
	    bucket: string;
	    prefix: string;
	    driveId?: string;
	    studyOid?: string;
	    url: string;
	    urls: string[];
	    createdAt: string;
	    expiresAt: string;
	    remaining: string;
	
	    static createFrom(source: any = {}) {
	        return new LANShareDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.name = source["name"];
	        this.provider = source["provider"];
	        this.bucket = source["bucket"];
	        this.prefix = source["prefix"];
	        this.driveId = source["driveId"];
	        this.studyOid = source["studyOid"];
	        this.url = source["url"];
	        this.urls = source["urls"];
	        this.createdAt = source["createdAt"];
	        this.expiresAt = source["expiresAt"];
	        this.remaining = source["remaining"];
	    }
	}
	export class ObjectDTO {
	    Key: string;
	    Name: string;
	    IsDir: boolean;
	    Size: number;
	    Type: string;
	    LastModified: string;
	
	    static createFrom(source: any = {}) {
	        return new ObjectDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Name = source["Name"];
	        this.IsDir = source["IsDir"];
	        this.Size = source["Size"];
	        this.Type = source["Type"];
	        this.LastModified = source["LastModified"];
	    }
	}
	export class ObjectListDTO {
	    Items: ObjectDTO[];
	    Truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ObjectListDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Items = this.convertValues(source["Items"], ObjectDTO);
	        this.Truncated = source["Truncated"];
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
	export class ProxyDTO {
	    Mode: string;
	    Host: string;
	    Port: number;
	    Username: string;
	    Password: string;
	    NoProxy: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Mode = source["Mode"];
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.Username = source["Username"];
	        this.Password = source["Password"];
	        this.NoProxy = source["NoProxy"];
	    }
	}
	export class ProxyModeDTO {
	    Mode: string;
	    Label: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyModeDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Mode = source["Mode"];
	        this.Label = source["Label"];
	    }
	}
	export class RWSDatasetFieldDTO {
	    OID: string;
	    Name: string;
	    DataType: string;
	    Length: number;
	
	    static createFrom(source: any = {}) {
	        return new RWSDatasetFieldDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OID = source["OID"];
	        this.Name = source["Name"];
	        this.DataType = source["DataType"];
	        this.Length = source["Length"];
	    }
	}
	export class RWSDatasetDTO {
	    OID: string;
	    Name: string;
	    Domain: string;
	    Type: string;
	    Fields: RWSDatasetFieldDTO[];
	
	    static createFrom(source: any = {}) {
	        return new RWSDatasetDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OID = source["OID"];
	        this.Name = source["Name"];
	        this.Domain = source["Domain"];
	        this.Type = source["Type"];
	        this.Fields = this.convertValues(source["Fields"], RWSDatasetFieldDTO);
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
	export class RWSDatasetDataDTO {
	    Headers: string[];
	    Rows: string[][];
	    Type: string;
	
	    static createFrom(source: any = {}) {
	        return new RWSDatasetDataDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Headers = source["Headers"];
	        this.Rows = source["Rows"];
	        this.Type = source["Type"];
	    }
	}
	
	export class RWSSiteDTO {
	    OID: string;
	    Name: string;
	    StudySiteNumber: string;
	
	    static createFrom(source: any = {}) {
	        return new RWSSiteDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OID = source["OID"];
	        this.Name = source["Name"];
	        this.StudySiteNumber = source["StudySiteNumber"];
	    }
	}
	export class RWSStudyDTO {
	    OID: string;
	    Name: string;
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new RWSStudyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OID = source["OID"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	    }
	}
	export class RWSSubjectDTO {
	    SubjectKey: string;
	    SiteRef: string;
	    Status: string;
	
	    static createFrom(source: any = {}) {
	        return new RWSSubjectDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SubjectKey = source["SubjectKey"];
	        this.SiteRef = source["SiteRef"];
	        this.Status = source["Status"];
	    }
	}
	export class SPItemDTO {
	    ID: string;
	    Name: string;
	    IsFolder: boolean;
	    Size: number;
	    CreatedTime: string;
	    LastModified: string;
	    CreatedBy: string;
	    LastModifiedBy: string;
	    DownloadURL: string;
	
	    static createFrom(source: any = {}) {
	        return new SPItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.IsFolder = source["IsFolder"];
	        this.Size = source["Size"];
	        this.CreatedTime = source["CreatedTime"];
	        this.LastModified = source["LastModified"];
	        this.CreatedBy = source["CreatedBy"];
	        this.LastModifiedBy = source["LastModifiedBy"];
	        this.DownloadURL = source["DownloadURL"];
	    }
	}
	export class SPTransferItemDTO {
	    ID: string;
	    Name: string;
	    ItemID: string;
	    LocalPath: string;
	
	    static createFrom(source: any = {}) {
	        return new SPTransferItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.ItemID = source["ItemID"];
	        this.LocalPath = source["LocalPath"];
	    }
	}
	export class SearchTargetDTO {
	    Bucket: string;
	    Dir: string;
	    Keyword: string;
	    IsSearch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SearchTargetDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Bucket = source["Bucket"];
	        this.Dir = source["Dir"];
	        this.Keyword = source["Keyword"];
	        this.IsSearch = source["IsSearch"];
	    }
	}
	export class ShareSessionDTO {
	    Name: string;
	    Provider: string;
	    Bucket: string;
	    DriveID: string;
	    DriveName: string;
	    ItemPath: string;
	    ExpiresAt: string;
	    Remaining: string;
	    Expired: boolean;
	    ReadOnly: boolean;
	    Account: AccountDTO;
	
	    static createFrom(source: any = {}) {
	        return new ShareSessionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Provider = source["Provider"];
	        this.Bucket = source["Bucket"];
	        this.DriveID = source["DriveID"];
	        this.DriveName = source["DriveName"];
	        this.ItemPath = source["ItemPath"];
	        this.ExpiresAt = source["ExpiresAt"];
	        this.Remaining = source["Remaining"];
	        this.Expired = source["Expired"];
	        this.ReadOnly = source["ReadOnly"];
	        this.Account = this.convertValues(source["Account"], AccountDTO);
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
	export class TransferItemDTO {
	    ID: string;
	    Name: string;
	    Key: string;
	    LocalPath: string;
	
	    static createFrom(source: any = {}) {
	        return new TransferItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Key = source["Key"];
	        this.LocalPath = source["LocalPath"];
	    }
	}
	export class TransferResultDTO {
	    Total: number;
	    Success: number;
	    Failed: number;
	    Errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new TransferResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Total = source["Total"];
	        this.Success = source["Success"];
	        this.Failed = source["Failed"];
	        this.Errors = source["Errors"];
	    }
	}

}

