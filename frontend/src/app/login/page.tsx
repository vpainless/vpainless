"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card"
import {
	Tabs,
	TabsContent,
	TabsList,
	TabsTrigger,
} from "@/components/ui/tabs"
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { useRouter } from "next/navigation"
import { User, UsersService } from "@/lib/api"
import * as Auth from "@/lib/auth-context"
import type { User as AuthUser } from "@/lib/auth-context"
import { toast } from "sonner"

export default function LoginPage() {
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");

	const { setUser } = Auth.useAuth();
	const router = useRouter();


	const handleLogin = function() {
		const token = btoa(`${username}:${password}`);
		const user: AuthUser = {
			id: undefined,
			name: username,
			group: undefined,
			token: token,
			role: undefined,
		}
		setUser(user);

		UsersService.getMe()
			.then(result => {
				user.group = result.group_id;
				user.role = result.role;
				user.id = result.id;
				setUser(user);
				toast("Login Successful", {
					duration: 4000,
				})

				if (user.role === User.role.CLIENT && user.group) {
					router.replace("/client");
				} else {
					router.replace("/admin");
				}
			})
			.catch(resp => {
				toast.error("Login Failed", {
					description: resp.response.data.error,
					duration: 4000,
				})
			});
	};

	const handleRegister = function() {
		setUser(null);
		const user: User = {
			username: username,
			password: password,
		}
		UsersService.postUser(user)
			.then(() => {
				handleLogin()
			})
			.catch(resp => {
				toast.error("Register Failed", {
					description: resp.response.data.error,
					duration: 4000,
				})
			});
	};

	const handleLoginKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
		if (e.key !== 'Enter') {
			return;
		}
		handleLogin();
	};
	const handleRegisterKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
		if (e.key !== 'Enter') {
			return;
		}
		handleRegister();
	};

	return (
		<main className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-mono)]">
			<div className="flex flex-col row-start-2 w-full max-w-sm space-y-4">
				<Tabs defaultValue="login">
					<TabsList className="grid w-full grid-cols-2">
						<TabsTrigger value="login">Login</TabsTrigger>
						<TabsTrigger value="register">Register</TabsTrigger>
					</TabsList>
					<TabsContent value="login" className="min-h-[350px]">
						<Card className="w-full">
							<CardHeader>
								<CardTitle>Login</CardTitle>
								<CardDescription>
									Enter you credentials and click on login button.
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4">
								<div className="space-y-1">
									<Label htmlFor="username">Username</Label>
									<Input
										id="username"
										value={username}
										onChange={(e) => setUsername(e.target.value)}
									/>
								</div>
								<div className="space-y-1">
									<Label htmlFor="password">Password</Label>
									<Input
										id="password"
										type="password"
										value={password}
										onChange={(e) => setPassword(e.target.value)}
										onKeyDown={handleLoginKeyDown}
									/>
								</div>
								<Button className="w-full" onClick={handleLogin}>
									Login
								</Button>
							</CardContent>
						</Card>
					</TabsContent>
					<TabsContent value="register" className="min-h-[350px]">
						<Card className="w-full">
							<CardHeader>
								<CardTitle>Register</CardTitle>
								<CardDescription>
									Pick a unique user name and click on the register. You can create a user group later.
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4">
								<div className="space-y-1">
									<Label htmlFor="register">Username</Label>
									<Input
										id="register"
										value={username}
										onChange={(e) => setUsername(e.target.value)}
									/>
								</div>
								<div className="space-y-1">
									<Label htmlFor="password">Password</Label>
									<Input
										id="password"
										type="password"
										value={password}
										onChange={(e) => setPassword(e.target.value)}
										onKeyDown={handleRegisterKeyDown}
									/>
								</div>
								<Button className="w-full" onClick={handleRegister}>
									Register
								</Button>
							</CardContent>
						</Card>
					</TabsContent>
				</Tabs>
			</div >
		</main >
	);
}
