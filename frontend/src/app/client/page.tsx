"use client";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useRouter } from "next/navigation"
import { useEffect, useState } from "react";
import { Instance, InstancesService } from "@/lib/api"
import { useAuth } from "@/lib/auth-context"
import { toast } from "sonner"
import { Loader2 } from "lucide-react";
import {
	Drawer,
	DrawerClose,
	DrawerContent,
	DrawerDescription,
	DrawerFooter,
	DrawerHeader,
	DrawerTitle,
	DrawerTrigger,
} from "@/components/ui/drawer"

import { AlertCircleIcon } from "lucide-react"
import {
	Alert,
	AlertDescription,
	AlertTitle,
} from "@/components/ui/alert"

interface ErrorResponse {
	response?: {
		data?: {
			error?: string;
		};
	};
}

export default function ClientPage() {
	const { user } = useAuth();
	const router = useRouter();

	useEffect(function() {
		if (!user) {
			router.replace("/login");
		}
	}, [user]);

	if (!user) return null;

	const [instance, setInstance] = useState<Instance>();
	const [status, setStatus] = useState<string>();
	const [isPolling, setIsPolling] = useState<boolean>(false);

	useEffect(function() {
		InstancesService.listInstances()
			.then((instances) => {
				if (!instances) {
					return;
				}
				if (instances.length > 1) {
					toast.error("Listing insatnces failed!", {
						description: "You have more that one instance",
						duration: 4000,
					})
					return
				}

				setInstance(instances.at(0));
			})
			.catch(err => {
				const e = err as ErrorResponse;
				toast.error("Listing insatnces failed!", {
					description: e.response?.data?.error || "unknown error",
					duration: 4000,
				})
			});
	}, [])

	const handleCopy = async () => {
		try {
			if (!instance) {
				return;
			}
			await navigator.clipboard.writeText(instance.connection_string ?? "");
			toast("Connection string copied to the clipboard!", {
				duration: 4000,
			})
		} catch {
			toast.error("Could not copy connection string to the clipboard!", {
				description: "Try manually copying the connection string text above",
				duration: 4000,
			})
		}
	};


	const poll = async (instance: Instance) => {
		if (isPolling || !instance) {
			return;
		}
		setIsPolling(true);
		const workload = async () => {
			try {
				const updatedInstance = await InstancesService.getInstance(instance.id!);
				setInstance(updatedInstance);

				const statusStr = mapStatus(updatedInstance.status!);
				setStatus(statusStr);
				if (!statusStr) {
					clearInterval(id);
					toast.success("VPN Ready!", {
						duration: 4000,
					});
					setIsPolling(false);
				}
			} catch (err: unknown) {
				const e = err as ErrorResponse;
				toast.error("Error refreshing instance!", {
					description: e.response?.data?.error || "unknown error",
					duration: 4000,
				});
			}
		}
		const id = setInterval(workload, 5000);
		workload();
	}

	const handleCreate = async () => {
		if (instance) {
			try {
				setStatus("deleting...");
				await InstancesService.deleteInstance(instance.id!);
				setInstance(undefined);
			} catch (err: unknown) {
				setStatus(undefined)
				const e = err as ErrorResponse;
				toast.error("Error deleting instance!", {
					description: e.response?.data?.error || "unknown error",
					duration: 4000,
				});
				return;
			};
		}

		try {
			setStatus(mapStatus(Instance.status.OFF));
			const instance = await InstancesService.postInstance();
			setInstance(instance);
			poll(instance);
			return;
		} catch (err: unknown) {
			const e = err as ErrorResponse;
			setStatus(undefined)
			toast.error("Error creating instance!", {
				description: e.response?.data?.error || "unknown error",
				duration: 4000,
			});
		};
	};

	const renderClient = function() {
		const renewButton = function() {
			if (status) {
				return (
					<div className="space-y-4">
						<Button className="w-full" variant="default" disabled>
							<Loader2 className="h-4 w-4 animate-spin" /> {status}
						</Button>
						<Alert>
							<AlertCircleIcon />
							<AlertTitle>Please be patient!</AlertTitle>
							<AlertDescription>
								<p>Instance creation may take 5 minute or more.</p>
							</AlertDescription>
						</Alert>
					</div>
				)
			}

			if (!instance) {
				return (
					<Button variant="default" onClick={handleCreate}> Create </Button>
				)
			}

			return (
				<Drawer>
					<DrawerTrigger asChild>
						<Button className="w-full">Renew</Button>
					</DrawerTrigger>
					<DrawerContent>
						<DrawerHeader>
							<DrawerTitle>Are you sure?</DrawerTitle>
							<DrawerDescription>This action cannot be undone.</DrawerDescription>
						</DrawerHeader>
						<DrawerFooter>
							<Button className="w-full" onClick={handleCreate}>Renew</Button>
							<DrawerClose asChild>
								<Button className="w-full" variant="outline">Cancel</Button>
							</DrawerClose>
						</DrawerFooter>
					</DrawerContent>
				</Drawer>
			)
		}

		const connection_string = instance?.connection_string;
		if (instance && instance.status != Instance.status.OK) {
			poll(instance);
		};

		return (
			<div className="flex flex-col row-start-2 w-full max-w-sm space-y-4">
				<h1 className="text-2xl font-bold text-center">VPN</h1>
				{connection_string && (<Textarea readOnly value={connection_string}></Textarea>)}
				{connection_string && <Button variant="secondary" onClick={handleCopy}>Copy</Button>}
				{renewButton()}
			</div >
		)
	}

	const renderAdmin = function() {
		return (
			<div className="flex flex-col row-start-2 w-full max-w-sm space-y-4">
				Admin Panel Not implemented yet! Come back later!
			</div>
		)
	}

	return (
		<main className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-4 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-mono)]">
			{user.role === "client" ? renderClient() : renderAdmin()}
		</main >
	);
}

function mapStatus(status: Instance.status): string | undefined {
	switch (status) {
		case Instance.status.INITIALIZING:
			return "initializing...";
		case Instance.status.OK:
			return undefined;
		default:
			return "creating...";
	}
}

