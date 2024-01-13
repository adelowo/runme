/* eslint-disable */
// @generated by protobuf-ts 2.9.3 with parameter output_javascript,optimize_code_size,long_type_string,add_pb_suffix,ts_nocheck,eslint_disable
// @generated from protobuf file "runme/runner/v2alpha1/runner.proto" (package "runme.runner.v2alpha1", syntax proto3)
// tslint:disable
// @ts-nocheck
import { RunnerService } from "./runner_pb";
import { stackIntercept } from "@protobuf-ts/runtime-rpc";
/**
 * @generated from protobuf service runme.runner.v2alpha1.RunnerService
 */
export class RunnerServiceClient {
    constructor(_transport) {
        this._transport = _transport;
        this.typeName = RunnerService.typeName;
        this.methods = RunnerService.methods;
        this.options = RunnerService.options;
    }
    /**
     * Execute executes a program. Examine "ExecuteRequest" to explore
     * configuration options.
     *
     * It's a bidirectional stream RPC method. It expects the first
     * "ExecuteRequest" to contain details of a program to execute.
     * Subsequent "ExecuteRequest" should only contain "input_data" as
     * other fields will be ignored.
     *
     * @generated from protobuf rpc: Execute(stream runme.runner.v2alpha1.ExecuteRequest) returns (stream runme.runner.v2alpha1.ExecuteResponse);
     */
    execute(options) {
        const method = this.methods[0], opt = this._transport.mergeOptions(options);
        return stackIntercept("duplex", this._transport, method, opt);
    }
}
