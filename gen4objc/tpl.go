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

@interface {{title .}} : NSObject
{{range .Fields}}
{{fieldDecl .}}
{{end}}
+ (instancetype)init;
- (void) marshal:(id<GSWriter>) writer;
- (void) unmarshal:(id<GSReader>) reader;
{{if isException .}}
- (NSError*) asNSError;
{{end}}
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

{{if isPOD .}}

- (void) marshal:(id<GSWriter>) writer {
{{range .Fields}}
{{marshalField .}}
{{end}}
}

- (void) unmarshal:(id<GSReader>) reader {
{{range .Fields}}
    {
        {{unmarshalField .}}
    }
{{end}}
}

{{else}}
- (void) marshal:(id<GSWriter>) writer {
    [writer WriteByte :(UInt8){{len .Fields}}];
{{range .Fields}}
    [writer WriteByte :(UInt8){{tagValue .Type}}];
{{marshalField .}}
{{end}}
}
- (void) unmarshal:(id<GSReader>) reader {

    UInt8 __fields = [reader ReadByte];

{{range .Fields}}
    {
        UInt8 tag = [reader ReadByte];

        if(tag != GSTagSkip) {
        {{unmarshalField .}}
        }

        if(-- __fields == 0) {
            return;
        }
    }
{{end}}

    for(int i = 0; i < (int)__fields; i ++) {
        UInt8 tag = [reader ReadByte];

        if (tag == GSTagSkip) {
            continue;
        }

        [reader ReadSkip:tag];
    }
}
{{end}}

{{if isException .}}
- (NSError*) asNSError {
    NSString *domain = @"{{title .}}";

    NSDictionary *userInfo = @{ @"source" : self };

    NSError *error = [NSError errorWithDomain:domain code:-101 userInfo:userInfo];

    return error;
}
{{end}}

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
        case {{.ID}}:{
{{range .Params}}{{unmarshalParam . "call"}}{{end}}
            {{methodCall .}}
            {{if isAsync . | not}}
            GSResponse * callreturn  = [GSResponse init];
            callreturn.ID = call.ID;
            {{if notVoid .Return}}
        {{marshalReturn .Return}}
            {{end}}
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

    {{if isAsync .| not}}
    return GSCreatePromise(_channel,call,^id<GSPromise>(GSResponse* response,id block,NSError **error){

        if(response.Exception != (SInt8)-1) {
            switch(response.Exception){
            {{range .Exceptions}}
                case {{.ID}}:{
{{unmarshalReturn .Type 5}}
                    *error = [callreturn asNSError];
                    break;
                    }{{end}}
                default:{
                    NSString *domain = @"GSRemoteException";
                    *error = [NSError errorWithDomain:domain code:-101 userInfo:nil];
                }
            }

            return nil;
        }

{{if notVoid .Return}}{{unmarshalReturn .Return 2}}{{end}}
        return {{callback .}};
    });
    {{else}}
    return [_channel Post: call];
    {{end}}
}
{{end}}
@end



{{end}}


`
