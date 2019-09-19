--
-- PostgreSQL database dump
--

-- Dumped from database version 10.10 (Ubuntu 10.10-0ubuntu0.18.04.1)
-- Dumped by pg_dump version 10.10 (Ubuntu 10.10-0ubuntu0.18.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: friend_requests; Type: TABLE; Schema: public; Owner: buffalo
--

CREATE TABLE public.friend_requests (
    id uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    from_id uuid NOT NULL,
    to_id uuid NOT NULL,
    message text NOT NULL,
    status character varying(10) DEFAULT 'PENDING'::character varying NOT NULL
);


ALTER TABLE public.friend_requests OWNER TO buffalo;

--
-- Name: friendships; Type: TABLE; Schema: public; Owner: buffalo
--

CREATE TABLE public.friendships (
    created_at timestamp without time zone DEFAULT LOCALTIMESTAMP NOT NULL,
    user_id uuid NOT NULL,
    friend_id uuid NOT NULL
);


ALTER TABLE public.friendships OWNER TO buffalo;

--
-- Name: reports; Type: TABLE; Schema: public; Owner: buffalo
--

CREATE TABLE public.reports (
    id uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    by_id uuid NOT NULL,
    about_id uuid NOT NULL,
    info text NOT NULL
);


ALTER TABLE public.reports OWNER TO buffalo;

--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: buffalo
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO buffalo;

--
-- Name: users; Type: TABLE; Schema: public; Owner: buffalo
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    login character varying(255) NOT NULL,
    info character varying(255) NOT NULL,
    admin boolean NOT NULL
);


ALTER TABLE public.users OWNER TO buffalo;

--
-- Name: friend_requests friend_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.friend_requests
    ADD CONSTRAINT friend_requests_pkey PRIMARY KEY (id);


--
-- Name: friendships friendships_pkey; Type: CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT friendships_pkey PRIMARY KEY (user_id, friend_id);


--
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: buffalo
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: friend_requests friend_requests_users_from_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.friend_requests
    ADD CONSTRAINT friend_requests_users_from_id_fk FOREIGN KEY (from_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: friend_requests friend_requests_users_to_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.friend_requests
    ADD CONSTRAINT friend_requests_users_to_id_fk FOREIGN KEY (to_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: friendships friendships_users_friend_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT friendships_users_friend_id_fk FOREIGN KEY (friend_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: friendships friendships_users_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT friendships_users_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reports reports_users_about_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_users_about_id_fk FOREIGN KEY (about_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reports reports_users_by_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_users_by_id_fk FOREIGN KEY (by_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

