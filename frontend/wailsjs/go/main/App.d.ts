// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {eve} from '../models';

export function CheckAuth():Promise<boolean>;

export function GetLocation():Promise<eve.LocationResponse>;

export function GetRegisteredCharacters():Promise<Array<eve.ESIAuth>>;

export function GetZkill(arg1:number):Promise<Array<eve.FrontendKillmail>>;

export function LogOut():Promise<void>;

export function OpenAuth():Promise<void>;

export function SwitchCurrentCharacter(arg1:string):Promise<void>;
