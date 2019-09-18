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
-- Name: reports; Type: TABLE; Schema: public; Owner: buffalo
--

CREATE TABLE public.reports (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    subject_id uuid NOT NULL,
    message text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
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
    login character varying(255) NOT NULL,
    info character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.users OWNER TO buffalo;

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
-- Name: reports_subject_id_idx; Type: INDEX; Schema: public; Owner: buffalo
--

CREATE INDEX reports_subject_id_idx ON public.reports USING btree (subject_id);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: buffalo
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: reports reports_users_subject_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_users_subject_id_fk FOREIGN KEY (subject_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reports reports_users_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: buffalo
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_users_user_id_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

