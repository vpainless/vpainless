"use client";

import { Button } from "@/components/ui/button"
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
	HoverCard,
	HoverCardContent,
	HoverCardTrigger,
} from "@/components/ui/hover-card"
import {
	Table,
	TableBody,
	TableCaption,
	TableCell,
	TableFooter,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table"
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
	DialogClose,
} from "@/components/ui/dialog"
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

import { Clipboard, Pencil, Trash2, Users, Computer, ChartNoAxesCombined } from "lucide-react"

import { useRouter } from "next/navigation"
import { useEffect, useState } from "react";

import { useAuth } from "@/lib/auth-context"
import type { User as AuthUser } from "@/lib/auth-context"

import { User, Group, GroupsService, UsersService, Instance, InstancesService } from "@/lib/api"
import { toast } from "sonner"

interface ErrorResponse {
	response?: {
		data?: {
			error?: string;
		};
	};
}

export default function AdminPage() {

	const { user: principal, setUser } = useAuth();
	const router = useRouter();

	useEffect(function() {
		if (!principal) {
			router.replace("/login");
		}
		if (principal?.role !== "admin" && principal?.group) {
			router.replace("/");
		}
	}, [principal]);

	if (!principal) return null;

	const [groupName, setGroupName] = useState("");
	const [apikey, setAPIKey] = useState("");
	const [activeTab, setActiveTab] = useState("users");
	const [users, setUsers] = useState<User[]>([]);
	const [instances, setInstances] = useState<Instance[]>([]);
	const [refresh, setRefresh] = useState(false);

	const createGroup = function() {
		const group: Group = {
			name: groupName,
			vps: {
				apikey: apikey,
				provider: Group.provider.VULTR,
			}
		}

		GroupsService.postGroup(group)
			.then(result => {
				const newUser: AuthUser = {
					id: principal?.id,
					name: principal?.name,
					group: result.id,
					token: principal?.token,
					role: User.role.ADMIN,
				}
				setUser(newUser);
				setRefresh((prev: boolean) => !prev)
			})
			.catch(err => {
				const e = err as ErrorResponse;
				toast.error("Group Creation Failed!", {
					description: e.response?.data?.error || "unknown error",
					duration: 4000,
				})
			});
	}

	useEffect(() => {
		if (!principal.group) {
			return;
		}
		UsersService.listUsers()
			.then(result => {
				setUsers(result.users ?? [])
			})
			.catch(err => {
				const e = err as ErrorResponse
				toast.error("Listing users failed", {
					description: e.response?.data?.error || "unknown error",
					duration: 4000,
				})
			});
	}, [refresh])

	useEffect(() => {
		if (!principal.group) {
			return;
		}
		InstancesService.listInstances()
			.then(result => {
				setInstances(result ?? [])
			})
			.catch(err => {
				const e = err as ErrorResponse
				toast.error("Listing instances failed", {
					description: e.response?.data?.error || "unknown error",
					duration: 4000,
				})
			});
	}, [refresh])

	const renderGroupCreation = function() {
		return (
			<div className="flex flex-col row-start-2 w-full max-w-sm space-y-4">
				<Card className="w-[350px]">
					<CardHeader>
						<CardTitle>Create Group</CardTitle>
						<CardDescription>Get started by creating a group. You can invite users later.</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="grid w-full items-center gap-4">
							<div className="flex flex-col space-y-1.5">
								<Label htmlFor="name">Name</Label>
								<Input id="name" placeholder="group name"
									onChange={(e) => setGroupName(e.target.value)}
								/>
							</div>
							<div className="flex flex-col space-y-1.5">
								<HoverCard>
									<HoverCardTrigger>
										<Label htmlFor="apikey">Vultr API Key</Label>
									</HoverCardTrigger>
									<HoverCardContent className="w-80">
										<div className="flex justify-between space-x-4">
											<div className="space-y-1">
												<h4 className="text-sm font-semibold">Info</h4>
												<p className="text-sm">
													Currently we only support Vultr as the host provider service
												</p>
												<div className="flex items-center pt-2">
													<span className="text-xs text-muted-foreground">
														To create an API key, in your Vultr portal, navigate to: <br></br> <br></br> Account &gt; API
													</span>
												</div>
											</div>
										</div>
									</HoverCardContent>
								</HoverCard>
								<Input id="apikey" placeholder="api_key"
									onChange={(e) => setAPIKey(e.target.value)}
								/>
							</div>
						</div>
					</CardContent>
					<CardFooter className="flex">
						<Button className="w-full" onClick={createGroup}>Create</Button>
					</CardFooter>
				</Card >
			</div >
		)
	}

	const [dialogUsername, setDialogUsername] = useState("");
	const [dialogPassword, setDialogPassword] = useState("");
	const [isDialogOpen, setIsDialogOpen] = useState(false);

	const createUser = async function() {
		setIsDialogOpen(false);
		const newUser: User = {
			username: dialogUsername,
			password: dialogPassword,
			group_id: principal.group,
			role: User.role.CLIENT,
		}
		try {
			await UsersService.postUser(newUser)
			setRefresh((prev: boolean) => !prev)
		} catch (err: unknown) {
			const e = err as ErrorResponse;
			toast.error("Error creating a new client!", {
				description: e.response?.data?.error || "unknown error",
				duration: 4000,
			});
		}
	}

	const renderUsersCard = function() {
		return (
			<Card>
				<CardHeader>
					<CardTitle>Users</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<Table>
						<TableCaption>List of users in the group.</TableCaption>
						<TableHeader>
							<TableRow>
								<TableHead className="w-[100px]">ID</TableHead>
								<TableHead>Name</TableHead>
								<TableHead>Role</TableHead>
								<TableHead className="text-right">Actions</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{users.map((user: User) => (
								<TableRow key={user.id}>
									<TableCell className="font-medium">{user.id}</TableCell>
									<TableCell>{user.username}</TableCell>
									<TableCell>{user.role}</TableCell>
									<TableCell className="text-right space-x-1">
										<Button variant="outline" size="icon" disabled><Pencil /></Button>
										<Button variant="destructive" size="icon" disabled><Trash2 /></Button>
									</TableCell>
								</TableRow>
							))}
						</TableBody>
						<TableFooter>
							<TableRow>
								<TableCell colSpan={3}>Total</TableCell>
								<TableCell className="text-right">{users.length}</TableCell>
							</TableRow>
						</TableFooter>
					</Table>
				</CardContent>
				<CardFooter className="flex justify-end">
					<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
						<DialogTrigger asChild>
							<Button>Create Client</Button>
						</DialogTrigger>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Create Client</DialogTitle>
								<DialogDescription>
									Create a new client
								</DialogDescription>
							</DialogHeader>
							<div className="space-y-4">
								<div className="space-y-1">
									<Label htmlFor="username" className="text-right">
										Username
									</Label>
									<Input id="username" placeholder="username"
										onChange={(e) => setDialogUsername(e.target.value)} />
								</div>
								<div className="space-y-1">
									<Label htmlFor="password" className="text-right">
										Password
									</Label>
									<Input id="password" placeholder="password" type="password"
										onChange={(e) => setDialogPassword(e.target.value)} />
								</div>
							</div>
							<DialogFooter>
								<DialogClose asChild>
									<Drawer>
										<DrawerTrigger asChild>
											<Button className="w-full">Create</Button>
										</DrawerTrigger>
										<DrawerContent>
											<DrawerHeader>
												<DrawerTitle>Are you sure?</DrawerTitle>
												<DrawerDescription>Createing a client</DrawerDescription>
											</DrawerHeader>
											<div className="p-4">
												<div><span className="font-bold">Username:</span> {dialogUsername}</div>
												<div><span className="font-bold">Password:</span> {dialogPassword}</div>
											</div>
											<DrawerFooter>
												<Button className="w-full" onClick={createUser}>Yes</Button>
												<DrawerClose asChild>
													<Button className="w-full" variant="outline">Cancel</Button>
												</DrawerClose>
											</DrawerFooter>
										</DrawerContent>
									</Drawer>
								</DialogClose>
							</DialogFooter>
						</DialogContent>
					</Dialog>
				</CardFooter>
			</Card >
		)
	}

	const renderInstancesCard = function() {
		return (
			<Card>
				<CardHeader>
					<CardTitle>Instances</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<Table>
						<TableCaption>List of instances.</TableCaption>
						<TableHeader>
							<TableRow>
								<TableHead className="w-[100px]">ID</TableHead>
								<TableHead>User</TableHead>
								<TableHead>IP</TableHead>
								<TableHead>State</TableHead>
								<TableHead className="text-right">Actions</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{instances.map((instance: Instance) => (
								<TableRow key={instance.id}>
									<TableCell className="font-medium">{instance.id}</TableCell>
									<TableCell>{(function() {
										let result;
										users.map((user) => {
											if (user.id == instance.owner) {
												result = user.username;
											}
										});
										return result;
									})()}</TableCell>
									<TableCell>{instance.ip}</TableCell>
									<TableCell>{instance.status}</TableCell>
									<TableCell className="text-right space-x-1">
										<Button variant="outline" size="icon" onClick={(async function() {
											try {
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
										})}><Clipboard /></Button>
										<Drawer>
											<DrawerTrigger asChild>
												<Button variant="destructive" size="icon"><Trash2 /></Button>
											</DrawerTrigger>
											<DrawerContent>
												<DrawerHeader>
													<DrawerTitle>Are you sure?</DrawerTitle>
													<DrawerDescription>Deleting {instance.id}</DrawerDescription>
												</DrawerHeader>
												<DrawerFooter>
													<DrawerClose asChild>
														<Button variant="destructive" className="w-full" onClick={(async function() {
															try {
																await InstancesService.deleteInstance(instance.id!);
																toast("Instance deleted successfully!", {
																	duration: 4000,
																})
																setRefresh((prev: boolean) => !prev);
															} catch (err: unknown) {
																const e = err as ErrorResponse;
																toast.error("Error deleting instance!", {
																	description: e.response?.data?.error || "unknown error",
																	duration: 4000,
																});
																return;
															};
														})

														}>Yes</Button>
													</DrawerClose>
													<DrawerClose asChild>
														<Button className="w-full" variant="outline">Cancel</Button>
													</DrawerClose>
												</DrawerFooter>
											</DrawerContent>
										</Drawer>
									</TableCell>
								</TableRow>
							))}
						</TableBody>
						<TableFooter>
							<TableRow>
								<TableCell colSpan={4}>Total</TableCell>
								<TableCell className="text-right">{instances.length}</TableCell>
							</TableRow>
						</TableFooter>
					</Table>
				</CardContent>
			</Card >
		)
	}

	const renderAdminPanel = function() {
		return (
			<div className="flex flex-col row-start-2 w-full space-y-4">
				<div className="grid grid-cols-[1fr_3fr] gap-x-4 min-h-[400px]">
					<Card>
						<CardHeader>
							<CardTitle>Welcome, {principal.name}!</CardTitle>
						</CardHeader>
						<CardContent className="space-y-4">
							<div className="justify-between">
								<div className="flex items-center pt-2 text-sm">
									Group:
									<span className="text-xs text-muted-foreground">
										test_group
									</span>
								</div>
							</div>
							<div className="justify-between space-y-4">
								<Button variant="default"
									className={`justify-start w-full hover:bg-gray-800 ${activeTab === 'users' ? 'hover:bg-gray-600 bg-gray-600 text-white' : 'bg-gray text-gray-300'}`}
									onClick={() => setActiveTab("users")}
								><Users />Users</Button>
								<Button variant="default"
									className={`justify-start w-full hover:bg-gray-800 ${activeTab === 'instances' ? 'hover:bg-gray-600 bg-gray-600 text-white' : 'bg-gray text-gray-300'}`}
									onClick={() => setActiveTab("instances")}
								><Computer />Instances</Button>
								<Button variant="default"
									className={`justify-start w-full hover:bg-gray-800 ${activeTab === 'analytics' ? 'hover:bg-gray-600 bg-gray-600 text-white' : 'bg-gray text-gray-300'}`}
									onClick={() => setActiveTab("analytics")}
								><ChartNoAxesCombined />Analytics</Button>
							</div>
						</CardContent>
					</Card>
					{activeTab === 'users' && renderUsersCard()}
					{activeTab === 'instances' && renderInstancesCard()}
				</div>
			</div>
		)
	}

	return (
		<main className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-mono)]">
			{!principal.group && renderGroupCreation()}
			{principal.group && renderAdminPanel()}
		</main>
	)
}
