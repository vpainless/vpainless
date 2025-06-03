"use client";

import { createContext, useContext, useState, useEffect } from "react";

export type User = {
	id: string | undefined,
	name: string | undefined,
	group: string | undefined,
	token: string | undefined,
	role: string | undefined,
};

type AuthContextType = {
	user: User | null;
	setUser: (token: User | null) => void;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

let currentUser: User | null = null;

export function getUser(): User | null {
	return currentUser;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const [user, setUserState] = useState<User | null>(null);

	const setUser = (user: User | null) => {
		currentUser = user;
		setUserState(user);
	};

	useEffect(() => {
		currentUser = user;
	}, [user]);

	return (
		<AuthContext.Provider value={{
			user: user,
			setUser: setUser,
		}}>
			{children}
		</AuthContext.Provider>
	);
}

export function useAuth() {
	const ctx = useContext(AuthContext);
	if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
	return ctx;
}
