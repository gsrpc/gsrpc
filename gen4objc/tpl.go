package gen4objc

var t4objc = `

{{define "enum_predecl"}}
typedef enum {{title .}}:{{enumType .}} {{title .}};
{{end}}

{{define "enum_header"}}

// {{title .}} enum
enum {{title .}}:{{enumType .}}{ {{enumFields .}} };

// {{title .}} enum marshal/unmarshal helper interface
@interface {{title .}}Helper : NSObject
+ (void) marshal:({{title .}}) val withWriter:(id<GSWriter>) writer;
+ ({{title .}}) unmarshal:(id<GSReader>) reader;
+ (NSString*) tostring :({{title .}})val;
@end

{{end}}

{{define "enum_source"}}
@implementation {{title .}}Helper

+ (void) marshal:({{title .}}) val withWriter:(id<GSWriter>) writer {
    {{enumWrite .}};
}

+ ({{title .}}) unmarshal:(id<GSReader>) reader {
    return ({{title .}}){{enumRead .}};
}

+ (NSString*) tostring:({{title .}})val {
    {{$Enum := title .}}
    switch(val)
    {
    {{range .Constants}}
    case {{$Enum}}{{title2 .Name}}:
       return @"{{$Enum}}{{title2 .Name}}";
    {{end}}
    default:
       return @"Unknown val";
   }
}

@end

{{end}}


{{define "table_predecl"}}
@class {{title .}};
{{end}}

{{define "table_header"}}

// {{title .}} generated by objrpc tools
@interface {{title .}} : NSObject
{{range .Fields}}
{{fieldDecl .}}
{{end}}
+ (instancetype)init;
- (void) marshal:(id<GSWriter>) writer;
- (void) unmarshal:(id<GSReader>) reader;
@end

{{end}}

{{define "table_source"}}
@implementation {{title .}}
+ (instancetype)init {
    return [[{{title .}} alloc] init];
}
- (instancetype)init{
    if (self = [super init]){
        {{range .Fields}}
        _{{title2 .Name}} = {{defaultVal .Type}};
        {{end}}
    }
    return self;
}
- (void) marshal:(id<GSWriter>) writer {
{{range .Fields}}
{{marshalField .}}
{{end}}
}
- (void) unmarshal:(id<GSReader>) reader {
{{range .Fields}}
{{unmarshalField .}}
{{end}}
}

@end
{{end}}

{{define "exception_predecl"}}
@class {{title .}};
{{end}}

{{define "exception_header"}}

@interface {{title .}} : NSObject
{{range .Fields}}
{{fieldDecl .}}
{{end}}
+ (instancetype)init;
- (void) marshal:(id<GSWriter>) writer;
- (void) unmarshal:(id<GSReader>) reader;
- (NSError*) asNSError;
@end

{{end}}

{{define "exception_source"}}
@implementation {{title .}}
+ (instancetype)init {
    return [[{{title .}} alloc] init];
}
- (instancetype)init{
    if (self = [super init]){
        {{range .Fields}}
        _{{title2 .Name}} = {{defaultVal .Type}};
        {{end}}
    }
    return self;
}
- (void) marshal:(id<GSWriter>) writer {
{{range .Fields}}
{{marshalField .}}
{{end}}
}
- (void) unmarshal:(id<GSReader>) reader {
{{range .Fields}}
{{unmarshalField .}}
{{end}}
}

- (NSError*) asNSError {
    NSString *domain = @"{{title .}}";

    NSDictionary *userInfo = @{ @"source" : self };

    NSError *error = [NSError errorWithDomain:domain code:-101 userInfo:userInfo];

    return error;
}

@end
{{end}}

{{define "contract_header"}}

//{{title .}} generate by objrpc
@protocol {{title .}}<NSObject>
{{range .Methods}}
{{methodDecl .}};
{{end}}
@end

// {{title .}}Service generate by objrpc
@interface {{title .}}Service : NSObject<GSDispatcher>
+ (instancetype) init:(id<{{title .}}>)service withID:(UInt16) serviceID;

@property(readonly) UInt16 ID;

- (GSResponse *)Dispatch:(GSRequest *)call;

@end


@interface {{title .}}RPC : NSObject
+ (instancetype) initRPC:(id<GSChannel>) channel withID:(UInt16) serviceID;
{{range .Methods}}
{{rpcMethodDecl .}};
{{end}}
@end

{{end}}

{{define "contract_source"}}

{{$Contract := title .}}
@implementation {{title .}}Service{
    id<{{title .}}> _service;
}
+ (instancetype) init:(id<{{title .}}>)service withID:(UInt16) serviceID {
    return [[{{title .}}Service alloc] init: service withID: serviceID];
}
- (instancetype) init:(id<{{title .}}>)service withID:(UInt16) serviceID {
    if(self = [super init]) {
        _service = service;
        _ID = serviceID;
    }
    return self;
}

- (GSResponse*) Dispatch:(GSRequest*)call {
    switch(call.Method){
    {{range .Methods}}
    case {{.ID}}:
    {
{{range .Params}}{{unmarshalParam . "call"}}{{end}}
        {{methodCall .}}
        {{if notVoid .Return}}
        GSResponse * callreturn  = [GSResponse init];
        {{marshalReturn .Return}}
        callreturn.ID = call.ID;
        callreturn.Service = call.Service;
        return callreturn;
        {{end}}
        break;
    }
    {{end}}
    }
    return nil;
}

@end


@implementation {{title .}}RPC {
    id<GSChannel> _channel;
    UInt16 _serviceID;
}
+ (instancetype) initRPC:(id<GSChannel>) channel withID:(UInt16) serviceID {
    return [[{{title .}}RPC alloc] initRPC: channel withID: serviceID];
}
- (instancetype) initRPC:(id<GSChannel>) channel withID:(UInt16) serviceID {
    if(self = [super init]) {
        _channel = channel;
        _serviceID = serviceID;
    }
    return self;
}

{{range .Methods}}
{{rpcMethodDecl .}}{
    GSRequest* call = [GSRequest init];
    call.Service = _serviceID;
    call.Method = (UInt16){{.ID}};
    {{if .Params}}
    NSMutableArray * params = [NSMutableArray array];
{{marshalParams .Params}}
    call.Params = params;
    {{end}}

    return GSCreatePromise(_channel,call,^id<GSPromise>(GSResponse* response,id block){

{{if notVoid .Return}}{{unmarshalReturn .Return}}{{end}}
        return {{callback .}};
    });
}
{{end}}
@end



{{end}}


`