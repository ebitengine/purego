package main

import (
	"log"

	"github.com/ebitengine/purego/objc"
)

func init() {
	var p *objc.Protocol

	// Begin Objective-C protocol definition for: NSApplicationDelegate
	p = objc.AllocateProtocol("NSApplicationDelegate")
	if p != nil { // only register if doesn't exist
		p.AddMethodDescription(objc.RegisterName("application:delegateHandlesKey:"), "B32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationWillFinishLaunching:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("application:continueUserActivity:restorationHandler:"), "B40@0:8@16@24@?32", false, true)
		p.AddMethodDescription(objc.RegisterName("application:didDecodeRestorableState:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:didFailToContinueUserActivityWithType:error:"), "v40@0:8@16@24@32", false, true)
		p.AddMethodDescription(objc.RegisterName("application:didFailToRegisterForRemoteNotificationsWithError:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:didReceiveRemoteNotification:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:didRegisterForRemoteNotificationsWithDeviceToken:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:didUpdateUserActivity:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:handlerForIntent:"), "@32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:openFile:"), "B32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:openFileWithoutUI:"), "B32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:openFiles:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:openTempFile:"), "B32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:openURLs:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:printFile:"), "B32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:printFiles:withSettings:showPrintPanels:"), "Q44@0:8@16@24@32B40", false, true)
		p.AddMethodDescription(objc.RegisterName("application:userDidAcceptCloudKitShareWithMetadata:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:willContinueUserActivityWithType:"), "B32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:willEncodeRestorableState:"), "v32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("application:willPresentError:"), "@32@0:8@16@24", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidBecomeActive:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidChangeOcclusionState:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidChangeScreenParameters:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidFinishLaunching:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidHide:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidResignActive:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidUnhide:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDidUpdate:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationDockMenu:"), "@24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationOpenUntitledFile:"), "B24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationProtectedDataDidBecomeAvailable:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationProtectedDataWillBecomeUnavailable:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationShouldAutomaticallyLocalizeKeyEquivalents:"), "B24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationShouldHandleReopen:hasVisibleWindows:"), "B28@0:8@16B24", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationShouldOpenUntitledFile:"), "B24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationShouldTerminate:"), "Q24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationShouldTerminateAfterLastWindowClosed:"), "B24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationSupportsSecureRestorableState:"), "B24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationWillBecomeActive:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationWillHide:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationWillResignActive:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationWillTerminate:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationWillUnhide:"), "v24@0:8@16", false, true)
		p.AddMethodDescription(objc.RegisterName("applicationWillUpdate:"), "v24@0:8@16", false, true)
		var adoptedProtocol *objc.Protocol
		adoptedProtocol = objc.GetProtocol("NSObject")
		if adoptedProtocol == nil {
			log.Fatalln("protocol 'NSObject' does not exist")
		}
		p.AddProtocol(adoptedProtocol)
		p.Register()
		// Finished protocol: NSApplicationDelegate
	}
}
