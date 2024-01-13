/* eslint-disable */
// @generated by protobuf-ts 2.9.3 with parameter output_javascript,optimize_code_size,long_type_string,add_pb_suffix,ts_nocheck,eslint_disable
// @generated from protobuf file "runme/runner/v2alpha1/runner.proto" (package "runme.runner.v2alpha1", syntax proto3)
// tslint:disable
// @ts-nocheck
import { MessageType } from "@protobuf-ts/runtime";
import { UInt32Value } from "../../../google/protobuf/wrappers_pb";
/**
 * @generated from protobuf message runme.runner.v2alpha1.Project
 */
export interface Project {
    /**
     * project root folder
     *
     * @generated from protobuf field: string root = 1;
     */
    root: string;
    /**
     * list of environment files to try and load
     * start with
     *
     * @generated from protobuf field: repeated string env_load_order = 2;
     */
    envLoadOrder: string[];
}
/**
 * @generated from protobuf message runme.runner.v2alpha1.Winsize
 */
export interface Winsize {
    /**
     * number of rows (in cells)
     *
     * @generated from protobuf field: uint32 rows = 1;
     */
    rows: number;
    /**
     * number of columns (in cells)
     *
     * @generated from protobuf field: uint32 cols = 2;
     */
    cols: number;
    /**
     * width in pixels
     *
     * @generated from protobuf field: uint32 x = 3;
     */
    x: number;
    /**
     * height in pixels
     *
     * @generated from protobuf field: uint32 y = 4;
     */
    y: number;
}
/**
 * document_path is a path to the document which contains
 * the cell to execute.
 *
 * If project is set, document_path should be relative to the project root.
 * Otherwise, it should be an absolute path.
 *
 * @generated from protobuf message runme.runner.v2alpha1.ExecuteRequest
 */
export interface ExecuteRequest {
    /**
     * If it's a relative path and project is not set, directory is used as a base.
     *
     * @generated from protobuf field: string document_path = 1;
     */
    documentPath: string;
    /**
     * project represents a project in which the document is located.
     *
     * @generated from protobuf field: optional runme.runner.v2alpha1.Project project = 2;
     */
    project?: Project;
    /**
     * @generated from protobuf oneof: block
     */
    block: {
        oneofKind: "blockId";
        /**
         * @generated from protobuf field: string block_id = 8;
         */
        blockId: string;
    } | {
        oneofKind: "blockName";
        /**
         * @generated from protobuf field: string block_name = 9;
         */
        blockName: string;
    } | {
        oneofKind: undefined;
    };
    /**
     * directory to execute the program in. If not set,
     * the current working directory is used.
     *
     * @generated from protobuf field: string directory = 3;
     */
    directory: string;
    /**
     * env is a list of additional environment variables
     * that will be injected to the executed program.
     * They will override any env from the project.
     *
     * @generated from protobuf field: repeated string env = 4;
     */
    env: string[];
    /**
     * input_data is a byte array that will be send as input
     * to the program.
     *
     * @generated from protobuf field: bytes input_data = 5;
     */
    inputData: Uint8Array;
    /**
     * stop requests the running process to be stopped.
     * It is allowed only in the consecutive calls.
     *
     * @generated from protobuf field: runme.runner.v2alpha1.ExecuteStop stop = 6;
     */
    stop: ExecuteStop;
    /**
     * sets pty winsize
     * has no effect in non-interactive mode
     *
     * @generated from protobuf field: optional runme.runner.v2alpha1.Winsize winsize = 7;
     */
    winsize?: Winsize;
    /**
     * interactive, if true, will allow to process input_data.
     * When no more data is expected, EOT (0x04) character
     * must be sent in input_data.
     *
     * @generated from protobuf field: bool interactive = 10;
     */
    interactive: boolean;
}
/**
 * @generated from protobuf message runme.runner.v2alpha1.ProcessPID
 */
export interface ProcessPID {
    /**
     * @generated from protobuf field: int64 pid = 1;
     */
    pid: string;
}
/**
 * @generated from protobuf message runme.runner.v2alpha1.ExecuteResponse
 */
export interface ExecuteResponse {
    /**
     * exit_code is sent only in the final message.
     *
     * @generated from protobuf field: google.protobuf.UInt32Value exit_code = 1;
     */
    exitCode?: UInt32Value;
    /**
     * stdout_data contains bytes from stdout since the last response.
     *
     * @generated from protobuf field: bytes stdout_data = 2;
     */
    stdoutData: Uint8Array;
    /**
     * stderr_data contains bytes from stderr since the last response.
     *
     * @generated from protobuf field: bytes stderr_data = 3;
     */
    stderrData: Uint8Array;
    /**
     * pid contains the process' PID
     * this is only sent once in an initial response for background processes.
     *
     * @generated from protobuf field: runme.runner.v2alpha1.ProcessPID pid = 4;
     */
    pid?: ProcessPID;
}
/**
 * @generated from protobuf enum runme.runner.v2alpha1.ExecuteStop
 */
export declare enum ExecuteStop {
    /**
     * @generated from protobuf enum value: EXECUTE_STOP_UNSPECIFIED = 0;
     */
    UNSPECIFIED = 0,
    /**
     * @generated from protobuf enum value: EXECUTE_STOP_INTERRUPT = 1;
     */
    INTERRUPT = 1,
    /**
     * @generated from protobuf enum value: EXECUTE_STOP_KILL = 2;
     */
    KILL = 2
}
declare class Project$Type extends MessageType<Project> {
    constructor();
}
/**
 * @generated MessageType for protobuf message runme.runner.v2alpha1.Project
 */
export declare const Project: Project$Type;
declare class Winsize$Type extends MessageType<Winsize> {
    constructor();
}
/**
 * @generated MessageType for protobuf message runme.runner.v2alpha1.Winsize
 */
export declare const Winsize: Winsize$Type;
declare class ExecuteRequest$Type extends MessageType<ExecuteRequest> {
    constructor();
}
/**
 * @generated MessageType for protobuf message runme.runner.v2alpha1.ExecuteRequest
 */
export declare const ExecuteRequest: ExecuteRequest$Type;
declare class ProcessPID$Type extends MessageType<ProcessPID> {
    constructor();
}
/**
 * @generated MessageType for protobuf message runme.runner.v2alpha1.ProcessPID
 */
export declare const ProcessPID: ProcessPID$Type;
declare class ExecuteResponse$Type extends MessageType<ExecuteResponse> {
    constructor();
}
/**
 * @generated MessageType for protobuf message runme.runner.v2alpha1.ExecuteResponse
 */
export declare const ExecuteResponse: ExecuteResponse$Type;
/**
 * @generated ServiceType for protobuf service runme.runner.v2alpha1.RunnerService
 */
export declare const RunnerService: any;
export {};
